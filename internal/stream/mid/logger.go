package mid

import (
	"context"
	"log"

	"github.com/nats-io/nats.go"

	"github.com/TonyPath/activity-ingestion-service/internal/foundation/natskit"
)

func Logger(log *log.Logger) natskit.MsgHandlerMid {

	m := func(handler natskit.MsgHandler) natskit.MsgHandler {

		h := func(ctx context.Context, msg *nats.Msg) error {

			log.Printf("consume %q message : %s ", msg.Subject, string(msg.Data))

			err := handler(ctx, msg)

			return err
		}

		return h
	}

	return m
}
