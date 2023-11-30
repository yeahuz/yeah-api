package cqrs

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var jstream jetstream.JetStream

type Message interface {
	Name() string
}
type SetupOpts struct {
	NatsURL       string
	NatsAuthToken string
	AwsKey        string
	AwsSecret     string
}

func Setup(opts SetupOpts) *nats.Conn {
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

	nc.QueueSubscribe("SendEmailCode", "emails", func(m *nats.Msg) {
		var cmd SendEmailCodeCommand
		if err := gob.NewDecoder(bytes.NewBuffer(m.Data)).Decode(&cmd); err != nil {
			fmt.Printf("Error decoding: %s\n", err)
		}
		out, err := sesClient.SendEmail(context.Background(), &ses.SendEmailInput{
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

		//TODO: handle errors properly.
		if err != nil {
			return
		}

		_ = out
		if err := Send(NewEmailCodeSentEvent(cmd.Email)); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
	})

	jstream = js

	return nc
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
