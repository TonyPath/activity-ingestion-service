package natskit

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

// Consumer is the entrypoint to configure a value of it
type Consumer struct {
	nc   *nats.Conn
	log  *log.Logger
	subs map[string]*nats.Subscription
	mw   []MsgHandlerMid
}

// NewConsumer creates a Consumer value that can handle a set of subscriptions
func NewConsumer(natsURL string, log *log.Logger, mw []MsgHandlerMid) (*Consumer, error) {
	nc, err := setupNATSConnection(natsURL)
	if err != nil {
		return nil, fmt.Errorf("creating NATS consumer: %w", err)
	}

	return &Consumer{
		nc:   nc,
		log:  log,
		subs: make(map[string]*nats.Subscription),
		mw:   mw,
	}, nil
}

// Stream starts consuming event messages and sets a handler function for a given subject and queue
func (c *Consumer) Stream(ctx context.Context, subj, queue string, handler MsgHandler, mw ...MsgHandlerMid) error {
	if c == nil {
		return errors.New("consumer nil")
	}

	handler = wrapMiddleware(handler, mw)

	handler = wrapMiddleware(handler, c.mw)

	h := func(m *nats.Msg) {
		if err := handler(ctx, m); err != nil {
			return
		}
	}

	sub, err := c.nc.QueueSubscribe(subj, queue, h)
	if err != nil {
		return fmt.Errorf("subscribe %q subject: %w", subj, err)
	}

	c.log.Printf("Waiting to consume %q messages", subj)

	c.subs[subj] = sub
	return nil
}

// Close is used to unsubscribe subscribers and close the connection
func (c *Consumer) Close() {
	if c == nil {
		return
	}

	defer c.nc.Close()

	if c.subs != nil {
		for subject, subscription := range c.subs {
			if err := subscription.Unsubscribe(); err != nil {
				c.log.Printf("unsubscribe %q subject error : %v", subject, err)
				continue
			}

			c.log.Printf("unsubscribe %q subject", subject)
		}
	}
}
