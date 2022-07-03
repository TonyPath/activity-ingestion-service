package mid

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"

	"github.com/nats-io/nats.go"

	"github.com/TonyPath/activity-ingestion-service/internal/foundation/natskit"
)

func Panic() natskit.MsgHandlerMid {

	m := func(handler natskit.MsgHandler) natskit.MsgHandler {

		h := func(ctx context.Context, msg *nats.Msg) (err error) {

			defer func() {
				if r := recover(); r != nil {
					trace := debug.Stack()
					err = fmt.Errorf("panic: %v, trace: %s", r, string(trace))
					log.Printf("%s", debug.Stack())
				}
			}()

			return handler(ctx, msg)
		}

		return h
	}

	return m
}
