# Activity Ingestion Microservice

[![go.mod Go version](https://img.shields.io/github/go-mod/go-version/TonyPath/activity-ingestion-service)](https://github.com/TonyPath/activity-ingestion-service)

***

# Main Components

### Consumer
Listens for messages coming from ```ActivityFired``` and ```SendActivityDigest``` subjects 
by creating two subscriptions through ```Stream``` method provided by the consumer.
Both subscription are created under the group name ```activity-ingestion-svc``` so that in case of horizontal scaling e
ach message will be randomly balanced to only one of the subscribers

```go
consumer, err := natskit.NewConsumer(natsURL, log)

h1 := func(ctx context.Context, msg *nats.Msg) error {...}
consumer.Stream(ctx, "ActivityFired", "activity-ingestion-svc", h1)

h2 := func(ctx context.Context, msg *nats.Msg) error { ... }
consumer.Stream(ctx, "SendActivityDigest", "activity-ingestion-svc", h2)
```

### Producer
Publish messages to ```ActivityDigest``` subject

```go
producer, err := natskit.NewProducer(natsURL, log)
producer.Publish("ActivityDigest", activityDigestMessage)
```
