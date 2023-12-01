package cqrs

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var jstream jetstream.JetStream

var streamNames = map[string][]string{
	"auth": {"auth.sendEmailCode", "auth.sendPhoneCode", "auth.emailCodeSent", "auth.phoneCodeSent", "auth.emailCodeSendFailed", "auth.phoneCodeSendFailed"},
}

type Message interface {
	Name() string
}

type SetupOpts struct {
	NatsURL       string
	NatsAuthToken string
	AwsKey        string
	AwsSecret     string
}

func Setup(ctx context.Context, opts SetupOpts) (func(), error) {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(opts.AwsKey, opts.AwsSecret, "")),
		config.WithRegion("eu-north-1"),
	)

	if err != nil {
		log.Fatal(err)
	}

	sesClient := ses.NewFromConfig(cfg)

	nc, err := nats.Connect(opts.NatsURL, nats.Token(opts.NatsAuthToken))

	if err != nil {
		log.Fatal(err)
	}

	js, err := jetstream.New(nc)

	if err != nil {
		log.Fatal(err)
	}

	consumeContexts := []jetstream.ConsumeContext{}

	for key, value := range streamNames {
		stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     key,
			Subjects: value,
		})

		if err != nil {
			return nil, err
		}

		cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
			Durable:   key,
			AckPolicy: jetstream.AckExplicitPolicy,
		})

		if err != nil {
			return nil, err
		}

		cc, err := cons.Consume(func(m jetstream.Msg) {
			switch m.Subject() {
			case "auth.sendEmailCode":
				{
					var cmd SendEmailCodeCommand
					if err := gob.NewDecoder(bytes.NewBuffer(m.Data())).Decode(&cmd); err != nil {
						fmt.Printf("Error decoding: %s\n", err)
					}

					_, err := sesClient.SendEmail(ctx, &ses.SendEmailInput{
						Destination: &types.Destination{
							ToAddresses: []string{
								cmd.Email,
							},
						},
						Source: aws.String("Needs <noreply@needs.uz>"),
						Message: &types.Message{
							Subject: &types.Content{
								Data: aws.String(fmt.Sprintf("Your code: %s", cmd.Code)),
							},
							Body: &types.Body{
								Text: &types.Content{Data: aws.String(cmd.Code)},
							},
						},
					})
					if err != nil {
						m.NakWithDelay(time.Second * 5)
					}
					m.Ack()
					Send(NewEmailCodeSentEvent(cmd.Email))
					break
				}
			case "auth.emailCodeSent":
				{
					var ev EmailCodeSentEvent
					if err := gob.NewDecoder(bytes.NewBuffer(m.Data())).Decode(&ev); err != nil {
						fmt.Printf("Error decoding: %s\n", err)
					}
					//TODO: here you can increment counters for realtime, fast analytics
					fmt.Printf("Email sent to: %s\n", ev.Email)
					break
				}
			default:
				break
			}
		})

		consumeContexts = append(consumeContexts, cc)
	}

	jstream = js

	return func() {
		for _, cc := range consumeContexts {
			cc.Stop()
		}
		nc.Drain()
	}, nil
}

func Send(message Message) error {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(message); err != nil {
		return err
	}

	_, err := jstream.Publish(context.Background(), message.Name(), buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
