package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "go.uber.org/automaxprocs"

	"github.com/TonyPath/activity-ingestion-service/internal/foundation/cache"
	"github.com/TonyPath/activity-ingestion-service/internal/foundation/natskit"
	"github.com/TonyPath/activity-ingestion-service/internal/repo/memdb"
	"github.com/TonyPath/activity-ingestion-service/internal/stream/handlers"
	"github.com/TonyPath/activity-ingestion-service/internal/stream/mid"
)

const (
	natsURL   = "nats://localhost:4222"
	queueName = "activity-ingestion-svc"
)

func main() {
	logger := log.New(os.Stdout, "DIGEST ACTIVITY SERVICE : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if err := run(logger); err != nil {
		logger.Println("main: error:", err)
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	ctx := context.Background()

	// Init consumer
	//
	log.Println("Initializing NATS consumer")

	mw := []natskit.MsgHandlerMid{mid.Logger(log), mid.Error(log), mid.Panic()}
	consumer, err := natskit.NewConsumer(natsURL, log, mw)
	if err != nil {
		return err
	}
	defer consumer.Close()

	// Init producer
	//
	log.Println("Initializing NATS produces")

	producer, err := natskit.NewProducer(natsURL, log)
	if err != nil {
		return err
	}
	defer producer.Close()

	// Init in memory cache
	//
	log.Println("Initializing in Memory Cache")

	imCache := cache.New(5)
	defer imCache.Close()

	// Init storage
	//
	log.Println("Initializing Storage")

	storage, err := memdb.New()
	if err != nil {
		return nil
	}

	// Init Activity message group handlers
	//
	log.Println("Initializing Activity Message Handlers")

	activityGroup := handlers.ActivityGroup{
		Producer: producer,
		Cache:    imCache,
		Storage:  storage,
		Log:      log,
	}

	// Start Streaming
	//
	log.Println("Start NATS Streaming")

	err = consumer.Stream(ctx, handlers.SubActivityFired, queueName, activityGroup.HandleActivityFired)
	if err != nil {
		return err
	}

	err = consumer.Stream(ctx, handlers.SubSendActivityDigest, queueName, activityGroup.HandleSendActivityDigest)
	if err != nil {
		return err
	}

	// Wait for termination, i.e: CTRL-C
	//
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	log.Println("Shutdown service")

	return nil
}
