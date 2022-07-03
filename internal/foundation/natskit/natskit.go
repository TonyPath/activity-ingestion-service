// Package natskit contains a tiny NATS framework wrapper
package natskit

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

func setupNATSConnection(natsURL string) (*nats.Conn, error) {
	opts := nats.Options{
		Url:            natsURL,
		AllowReconnect: true,
		MaxReconnect:   -1,
		ReconnectWait:  time.Second,
		Timeout:        5 * time.Second,
	}
	nc, err := opts.Connect()
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS at %q: %w", natsURL, err)
	}

	return nc, nil
}
