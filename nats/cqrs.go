package nats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	yeahapi "github.com/yeahuz/yeah-api"
)

type CQRSService struct {
	nc              *nats.Conn
	js              jetstream.JetStream
	consumeContexts []jetstream.ConsumeContext
	handlers        map[string]yeahapi.CQRSHandler
}

func NewCQRSService(ctx context.Context, config yeahapi.CQRSConfig) (*CQRSService, error) {
	nc, err := nats.Connect(config.NatsURL, nats.Token(config.NatsAuthToken))
	if err != nil {
		return nil, err
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	c := &CQRSService{
		js:       js,
		nc:       nc,
		handlers: make(map[string]yeahapi.CQRSHandler),
	}

	for name, subjects := range config.Streams {
		stream, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     name,
			Subjects: subjects,
		})

		if err != nil {
			return nil, err
		}

		cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
			Durable:   name,
			AckPolicy: jetstream.AckExplicitPolicy,
		})

		if err != nil {
			return nil, err
		}

		cc, err := cons.Consume(func(m jetstream.Msg) {
			handle := c.handlers[m.Subject()]
			if handle == nil {
				return
			}

			if err := handle(m); err != nil {
				fmt.Println(err)
				if err := m.NakWithDelay(time.Second * 5); err != nil {
					// TODO: something went wrong with nats
					return
				}
			}
			//TODO: handle error
			m.Ack()
		})

		c.consumeContexts = append(c.consumeContexts, cc)
	}

	return c, nil
}

func (c *CQRSService) Handle(subject string, handler yeahapi.CQRSHandler) {
	c.handlers[subject] = handler
}

func (c *CQRSService) Close() error {
	for _, cc := range c.consumeContexts {
		cc.Stop()
	}

	return c.nc.Drain()
}

func (c *CQRSService) Publish(ctx context.Context, message yeahapi.CQRSMessage) error {
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(message); err != nil {
		return err
	}

	_, err := c.js.Publish(ctx, message.Subject(), buf.Bytes())
	return err
}
