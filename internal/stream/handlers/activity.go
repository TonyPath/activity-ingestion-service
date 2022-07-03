package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"

	"github.com/TonyPath/activity-ingestion-service/internal/foundation/cache"
	"github.com/TonyPath/activity-ingestion-service/internal/foundation/natskit"
	"github.com/TonyPath/activity-ingestion-service/internal/models"
	"github.com/TonyPath/activity-ingestion-service/internal/repo/memdb"
)

const (
	SubActivityFired      = "ActivityFired"
	SubSendActivityDigest = "SendActivityDigest"
	PubActivityDigest     = "ActivityDigest"
)

type Severity int

const (
	Critical Severity = iota
	High
	Medium
	Low
)

var activitySeverityMapper = map[string]Severity{
	"DeleteData":     Critical,
	"ChangePassword": High,
	"InstallProgram": Medium,
	"Login":          Low,
}

type activityStorage interface {
	SaveActivityForUser(ctx context.Context, activityID string, userID string, changedAt string) error
	GetActivityByUserID(ctx context.Context, userID string) ([]memdb.Activity, error)
}

type activityCache interface {
	FindOrAdd(kid string) cache.ResultStatus
}

// ActivityGroup groups a set of messages handlers
type ActivityGroup struct {
	Producer *natskit.Producer
	Cache    activityCache
	Storage  activityStorage
	Log      *log.Logger
}

// HandleActivityFired processes incoming messages to ActivityFired subject
func (ag ActivityGroup) HandleActivityFired(ctx context.Context, msg *nats.Msg) error {
	if msg.Subject != SubActivityFired {
		return fmt.Errorf("unknown subject %q", msg.Subject)
	}

	var activityFiredMsg models.ActivityFiredMessage
	if err := natskit.Decode(msg, &activityFiredMsg); err != nil {
		return fmt.Errorf("decoding message: %w", err)
	}

	// deduplicate
	kid, _ := activityFiredMsg.HashCode()
	if rs := ag.Cache.FindOrAdd(kid); rs == cache.Found {
		return nil
	}

	var activitySeverity Severity
	if severity, ok := activitySeverityMapper[activityFiredMsg.ActivityID]; ok {
		activitySeverity = severity
	}

	if activitySeverity == Critical {
		err := ag.Storage.SaveActivityForUser(
			ctx,
			activityFiredMsg.ActivityID,
			activityFiredMsg.UserID,
			activityFiredMsg.CreatedAt)

		if err != nil {
			return fmt.Errorf("saving activity: %w", err)
		}
	}

	return nil
}

// HandleSendActivityDigest processes incoming messages to SendactivityDigest subject
func (ag ActivityGroup) HandleSendActivityDigest(ctx context.Context, msg *nats.Msg) error {
	if msg.Subject != SubSendActivityDigest {
		return fmt.Errorf("unknown subject %q", msg.Subject)
	}

	var sendactivityDigest models.SendActivityDigestMessage
	if err := natskit.Decode(msg, &sendactivityDigest); err != nil {
		return fmt.Errorf("decoding message data: %w", err)
	}

	// deduplicate
	kid := sendactivityDigest.UserID
	if rs := ag.Cache.FindOrAdd(kid); rs == cache.Found {
		return nil
	}

	err := ag.flushActivitiesByUser(ctx, sendactivityDigest.UserID)
	if err != nil {
		return fmt.Errorf("flushing user's activitys: %w", err)
	}

	return nil
}

func (ag ActivityGroup) flushActivitiesByUser(ctx context.Context, userID string) error {
	activityDigestMessage := models.ActivityDigestMessage{
		UserID: userID,
	}

	userActivities, err := ag.Storage.GetActivityByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, v := range userActivities {
		activityDigestMessage.Activities = append(activityDigestMessage.Activities, models.ActivityDetails{
			ActivityID: v.ActivityID,
			CreatedAt:  v.CreatedAt,
		})
	}

	err = ag.Producer.Publish(PubActivityDigest, activityDigestMessage)
	if err != nil {
		return fmt.Errorf("publishing message: %w", err)
	}

	return nil
}
