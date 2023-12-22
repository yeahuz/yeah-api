package cqrs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"auth": {
		"auth.sendEmailCode", "auth.sendPhoneCode", "auth.emailCodeSent", "auth.phoneCodeSent", "auth.emailCodeSendFailed", "auth.phoneCodeSendFailed",
		"auth.loginTokenRejected", "auth.loginTokenAccepted",
	},
}

type Message interface {
	Subject() string
}

type Sender interface {
	Send(ctx context.Context, message Message) error
}

type SetupOpts struct {
	NatsURL       string
	NatsAuthToken string
	SmsClient     *smsclient.Client
	SesClient     *ses.Client
}

type Client struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func Setup(ctx context.Context, opts SetupOpts) (func(), *Client, error) {
	nc, err := nats.Connect(opts.NatsURL, nats.Token(opts.NatsAuthToken))

	if err != nil {
		return nil, nil, err
	}

	js, err := jetstream.New(nc)

	if err != nil {
		return nil, nil, err
	}

	client := &Client{nc: nc, js: js}

	smsClient := opts.SmsClient
	sesClient := opts.SesClient

	consumeContexts := []jetstream.ConsumeContext{}

	for key, value := range streamNames {
		stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     key,
			Subjects: value,
		})

		if err != nil {
			return nil, nil, err
		}

		cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
			Durable:   key,
			AckPolicy: jetstream.AckExplicitPolicy,
		})

		if err != nil {
			return nil, nil, err
		}

		cc, err := cons.Consume(func(m jetstream.Msg) {
			switch m.Subject() {
			case "auth.sendEmailCode":
				{
					var cmd SendEmailCodeCommand
					if err := json.Unmarshal(m.Data(), &cmd); err != nil {
						fmt.Printf("Error unmarshalling: %s\n", err)
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
					client.Send(context.TODO(), NewEmailCodeSentEvent(cmd.Email))
					break
				}
			case "auth.sendPhoneCode":
				{
					var cmd SendPhoneCodeCommand
					if err := json.Unmarshal(m.Data(), &cmd); err != nil {
						fmt.Printf("Error unmarshalling: %s\n", err)
					}

					_, err := smsClient.Send(cmd.PhoneNumber[1:], fmt.Sprintf("Your verification code is %s. It expires in 15 minutes. Do not share this code! @needs.uz #%s", cmd.Code, cmd.Code))
					if err != nil {
						m.NakWithDelay(time.Second * 5)
						return
					}
					m.Ack()
					client.Send(context.TODO(), NewPhoneCodeSentEvent(cmd.PhoneNumber))
					break
				}
			case "auth.emailCodeSent":
				{
					var ev EmailCodeSentEvent
					if err := json.Unmarshal(m.Data(), &ev); err != nil {
						fmt.Printf("Error unmarshalling: %s\n", err)
					}
					if err := m.Ack(); err != nil {
						m.NakWithDelay(time.Second * 5)
					}
					break
				}
			case "auth.phoneCodeSent":
				{
					var ev PhoneCodeSentEvent
					if err := json.Unmarshal(m.Data(), &ev); err != nil {
						fmt.Printf("Error unmarshalling: %s\n", err)
					}
					if err := m.Ack(); err != nil {
						m.NakWithDelay(time.Second * 5)
					}
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
	}, client, nil
}

func (c *Client) Send(ctx context.Context, message Message) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(message); err != nil {
		return err
	}

	_, err := jstream.Publish(ctx, message.Subject(), buf.Bytes())
	return err
}
