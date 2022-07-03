package natskit

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

// Producer is the entrypoint to configure a value of it
type Producer struct {
	nc  *nats.Conn
	log *log.Logger
}

// NewProducer creates a Producer value that is responsible to publish a message against a subject
func NewProducer(natsURL string, log *log.Logger) (*Producer, error) {
	nc, err := setupNATSConnection(natsURL)
	if err != nil {
		return nil, fmt.Errorf("creating NATS consumer: %w", err)
	}

	return &Producer{
		nc:  nc,
		log: log,
	}, nil
}

// Publish publishes a message to provided subject
func (p *Producer) Publish(subj string, msg any) error {
	if p == nil {
		return errors.New("producer nil")
	}

	m, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("serializing message: %w", err)
	}

	err = p.nc.Publish(subj, m)
	if err != nil {
		return fmt.Errorf("sending message: %w", err)
	}

	p.log.Printf("produce %q message : %s", subj, m)

	return nil
}

// Close is used to flush buffered data to NATS and close the connection
func (p *Producer) Close() {
	if p == nil {
		return
	}

	defer p.nc.Close()

	_ = p.nc.Flush()
}
