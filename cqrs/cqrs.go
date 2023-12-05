package cqrs

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/yeahuz/yeah-api/smsclient"
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
	SmsClient     *smsclient.Client
	SesClient     *ses.Client
}

func Setup(ctx context.Context, opts SetupOpts) (func(), error) {

	nc, err := nats.Connect(opts.NatsURL, nats.Token(opts.NatsAuthToken))

	if err != nil {
		log.Fatal(err)
	}

	js, err := jetstream.New(nc)

	if err != nil {
		log.Fatal(err)
	}

	smsClient := opts.SmsClient
	sesClient := opts.SesClient

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
						return
					}
					m.Ack()
					Send(NewEmailCodeSentEvent(cmd.Email))
					break
				}
			case "auth.sendPhoneCode":
				{
					var cmd SendPhoneCodeCommand

					if err := gob.NewDecoder(bytes.NewBuffer(m.Data())).Decode(&cmd); err != nil {
						fmt.Printf("Error decoding: %s\n", err)
					}

					err := smsClient.Send(cmd.PhoneNumber[1:], fmt.Sprintf("Your verification code is %s. It expires in 15 minutes. Do not share this code! @needs.uz #%s", cmd.Code, cmd.Code))
					if err != nil {
						fmt.Printf("Error: %s\n", err)
						m.NakWithDelay(time.Second * 5)
						return
					}
					m.Ack()
					Send(NewPhoneCodeSentEvent(cmd.PhoneNumber))
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
			case "auth.phoneCodeSent":
				{
					var ev PhoneCodeSentEvent
					if err := gob.NewDecoder(bytes.NewBuffer(m.Data())).Decode(&ev); err != nil {
						fmt.Printf("Error decoding: %s\n", err)
					}
					//TODO: here you can increment counters for realtime, fast analytics
					fmt.Printf("Phone code sent to: %s\n", ev.PhoneNumber)
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

	_, err := jstream.Publish(context.TODO(), message.Name(), buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
