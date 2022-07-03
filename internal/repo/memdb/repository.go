package memdb

import (
	"context"

	"github.com/hashicorp/go-memdb"
)

const (
	tActivities = "activities"
)

type Activity struct {
	ActivityID string `db:"activity_id"`
	UserID     string `db:"user_id"`
	CreatedAt  string `db:"created_at"`
}

var schema = &memdb.DBSchema{
	Tables: map[string]*memdb.TableSchema{
		tActivities: {
			Name: tActivities,
			Indexes: map[string]*memdb.IndexSchema{
				"id": {
					Name:   "id",
					Unique: true,
					Indexer: &memdb.CompoundIndex{
						Indexes: []memdb.Indexer{
							&memdb.StringFieldIndex{Field: "ActivityID"},
							&memdb.StringFieldIndex{Field: "UserID"},
						},
					},
				},
				"activity_id": {
					Name:    "activity_id",
					Indexer: &memdb.StringFieldIndex{Field: "ActivityID"},
				},
				"user_id": {
					Name:    "user_id",
					Indexer: &memdb.StringFieldIndex{Field: "UserID"},
				},
				"created_at": {
					Name:    "created_at",
					Indexer: &memdb.StringFieldIndex{Field: "CreatedAt"},
				},
			},
		},
	},
}

type Repository struct {
	db *memdb.MemDB
}

func New() (*Repository, error) {
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}
	r := Repository{
		db: db,
	}
	return &r, nil
}

func (r *Repository) SaveActivityForUser(_ context.Context, activityID string, userID string, createdAt string) error {
	activity := Activity{
		ActivityID: activityID,
		UserID:     userID,
		CreatedAt:  createdAt,
	}

	txn := r.db.Txn(true)
	defer txn.Abort()
	err := txn.Insert(tActivities, activity)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (r *Repository) GetActivityByUserID(_ context.Context, userID string) ([]Activity, error) {
	activities := make([]Activity, 0)

	txn := r.db.Txn(false)
	it, err := txn.Get(tActivities, "user_id", userID)
	if err != nil {
		return nil, err
	}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		a := obj.(Activity)
		activities = append(activities, a)
	}

	return activities, nil
}
