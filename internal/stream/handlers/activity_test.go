package handlers

import (
	"context"
	"log"
	"sync"
	"testing"

	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TonyPath/activity-ingestion-service/internal/foundation/cache"
	"github.com/TonyPath/activity-ingestion-service/internal/foundation/natskit"
	"github.com/TonyPath/activity-ingestion-service/internal/models"
	"github.com/TonyPath/activity-ingestion-service/internal/repo/memdb"
)

func TestHandleActivityFired(t *testing.T) {
	opts := natsserver.DefaultTestOptions
	opts.Port = -1
	opts.Cluster.Name = "testing"
	s := natsserver.RunServer(&opts)
	defer s.Shutdown()

	c, err := natskit.NewConsumer(s.Addr().String(), log.Default(), nil)
	require.NoError(t, err)
	defer c.Close()

	p, err := natskit.NewProducer(s.Addr().String(), log.Default())
	require.NoError(t, err)
	defer p.Close()

	cc := cache.New(1)
	defer cc.Close()

	storage, err := memdb.New()
	require.NoError(t, err)

	activityGroup := ActivityGroup{
		Producer: p,
		Cache:    cc,
		Storage:  storage,
		Log:      log.Default(),
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	ctx, cnl := context.WithCancel(context.Background())
	defer cnl()
	err = c.Stream(ctx, SubActivityFired, "test", activityGroup.HandleActivityFired, func(h natskit.MsgHandler) natskit.MsgHandler {
		return func(ctx context.Context, msg *nats.Msg) error {
			defer wg.Done()
			return h(ctx, msg)
		}
	})
	require.NoError(t, err)

	deleteData := models.ActivityFiredMessage{
		ActivityID: "DeleteData",
		UserID:     "12",
		CreatedAt:  "2021-06-15T23:44:14.5117569Z",
	}
	err = p.Publish(SubActivityFired, deleteData)
	assert.NoError(t, err)

	changePassword := models.ActivityFiredMessage{
		ActivityID: "ChangePassword",
		UserID:     "12",
		CreatedAt:  "2021-06-15T23:44:14.5117569Z",
	}
	err = p.Publish(SubActivityFired, changePassword)
	assert.NoError(t, err)

	wg.Wait()

	activities, err := storage.GetActivityByUserID(context.TODO(), "12")
	require.NoError(t, err)
	require.Len(t, activities, 1)

}
