package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/nats-io/nats.go/jetstream"
	yeahapi "github.com/yeahuz/yeah-api"
)

type EmailService struct {
	ses     *ses.Client
	cqrssrv yeahapi.CQRSService
}

func NewEmailService(cfg aws.Config, cqrssrv yeahapi.CQRSService) *EmailService {
	ses := ses.NewFromConfig(cfg)
	return &EmailService{
		ses:     ses,
		cqrssrv: cqrssrv,
	}
}

func (e *EmailService) HandleSendEmailCode(m jetstream.Msg) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var cmd yeahapi.SendEmailCodeCmd
	if err := json.Unmarshal(m.Data(), &cmd); err != nil {
		return err
	}

	_, err := e.ses.SendEmail(ctx, &ses.SendEmailInput{
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
		return err
	}

	e.cqrssrv.Publish(ctx, yeahapi.NewEmailCodeSentEvent(cmd.Email))

	return nil
}
