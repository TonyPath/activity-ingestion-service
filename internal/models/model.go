package models

import (
	"fmt"
	"hash/fnv"
	"strconv"
)

type ActivityFiredMessage struct {
	ActivityID string `json:"ActivityID"`
	UserID     string `json:"UserID"`
	CreatedAt  string `json:"ChangedAt"`
}

func (activity ActivityFiredMessage) HashCode() (string, error) {
	s := fmt.Sprintf("%s-%s-%s", activity.ActivityID, activity.UserID, activity.CreatedAt)

	h := fnv.New64a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return "", fmt.Errorf("hashing ActivityFiredMessage: %w", err)
	}

	return strconv.FormatUint(h.Sum64(), 10), nil
}

type SendActivityDigestMessage struct {
	UserID string `json:"UserID"`
}

type ActivityDigestMessage struct {
	UserID     string            `json:"UserID"`
	Activities []ActivityDetails `json:"Activities"`
}

type ActivityDetails struct {
	ActivityID string `json:"ActivityID"`
	CreatedAt  string `json:"CreatedAt"`
}
