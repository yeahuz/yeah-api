package cqrs

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

var jstream jetstream.JetStream

func Setup(url string) *nats.Conn {
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatal(err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	jstream = js

	return nc
}

func Send(cmd Command) error {
	buf := &bytes.Buffer{}
	if err := gob.NewEncoder(buf).Encode(cmd); err != nil {
		return err
	}

	_, err := jstream.Publish(context.Background(), cmd.Name(), buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
