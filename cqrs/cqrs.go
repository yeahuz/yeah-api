package cqrs

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
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
	NatsURL        string
	NatsAuthToken  string
	AwsKey         string
	AwsSecret      string
	SmsApiBaseUrl  string
	SmsApiEmail    string
	SmsApiPassword string
}

type smsClient struct {
	email           string
	password        string
	baseUrl         string
	timeoutDuration time.Duration
	token           string
	refreshing      atomic.Bool
	cond            *sync.Cond
}

type requestOpts struct {
	url        string
	method     string
	dataReader io.Reader
}

type tokenData struct {
	Token string `json:"token"`
}

type TokenResponse struct {
	Message   string    `json:"message"`
	Data      tokenData `json:"data"`
	TokenType string    `json:"token_type"`
}

func (sc *smsClient) getToken(ctx context.Context) error {
	form := url.Values{}
	form.Add("email", sc.email)
	form.Add("password", sc.password)
	req, err := http.NewRequest("POST", sc.baseUrl+"/auth/login", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return err
	}

	sc.token = tokenResponse.Data.Token
	sc.cond.Signal()

	return nil
}

func (sc *smsClient) request(opts requestOpts) error {
	ctx, cancel := context.WithTimeout(context.Background(), sc.timeoutDuration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, opts.method, sc.baseUrl+opts.url, opts.dataReader)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", sc.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		if !sc.refreshing.Load() {
			sc.refreshing.Store(true)
			sc.cond.L.Lock()
			go sc.getToken(ctx)
		}
		sc.cond.Wait()
		sc.cond.L.Unlock()
		sc.refreshing.Store(false)
		return sc.request(opts)
	}

	if err != nil {
		return err
	}

	//TODO: properly handle errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request(): Request failed with status code: %d", resp.StatusCode)
	}

	return nil
}

func (sc *smsClient) send(to, message string) error {
	form := url.Values{}
	form.Add("message", message)
	form.Add("from", "4546")
	form.Add("mobile_phone", to)

	encoded := form.Encode()

	err := sc.request(requestOpts{
		url:        "/message/sms/send",
		method:     "POST",
		dataReader: strings.NewReader(encoded),
	})

	if err != nil {
		return err
	}

	return nil
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
	smsClient := smsClient{
		timeoutDuration: time.Second * 30,
		email:           opts.SmsApiEmail,
		password:        opts.SmsApiPassword,
		baseUrl:         opts.SmsApiBaseUrl,
		cond:            sync.NewCond(&sync.Mutex{}),
	}

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

					err := smsClient.send(cmd.PhoneNumber[1:], fmt.Sprintf("Your verification code is %s. It expires in 15 minutes. Do not share this code! @needs.uz #%s", cmd.Code, cmd.Code))
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
