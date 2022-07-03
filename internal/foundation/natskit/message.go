package natskit

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

// MsgHandler is a callback type that handles an incoming event message
type MsgHandler func(ctx context.Context, msg *nats.Msg) error

// Decode reads the raw data of an event message and decoded into the provided value
func Decode(msg *nats.Msg, val any) error {
	r := bytes.NewReader(msg.Data)
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(val); err != nil {
		return err
	}

	return nil
}
