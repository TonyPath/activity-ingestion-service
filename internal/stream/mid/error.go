package mid

import (
	"context"
	"log"

	"github.com/nats-io/nats.go"

	"github.com/TonyPath/activity-ingestion-service/internal/foundation/natskit"
)

func Error(log *log.Logger) natskit.MsgHandlerMid {

	m := func(handler natskit.MsgHandler) natskit.MsgHandler {

		h := func(ctx context.Context, msg *nats.Msg) error {

			if err := handler(ctx, msg); err != nil {
				log.Printf("ERROR: %v", err)
				return err
			}

			return nil
		}

		return h
	}

	return m
}
