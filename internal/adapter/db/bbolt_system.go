package db

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/mockctl-hq/mockctl/internal/core/ports"
	"go.etcd.io/bbolt"
)

var (
	bucketSettings  = []byte("settings_bucket")
	bucketAuth      = []byte("auth_bucket")
	bucketTelemetry = []byte("telemetry_bucket")
	bucketMetadata  = []byte("metadata_bucket")
)

// BBoltSystemStore implements ports.SystemStore using embedded bbolt.
type BBoltSystemStore struct {
	db       *bbolt.DB
	readOnly bool
}

// NewBBoltSystemStore initializes the system database with lock timeout fallback.
func NewBBoltSystemStore(dbPath string) (*BBoltSystemStore, error) {
	return newBBoltSystemStore(dbPath, false)
}

func newBBoltSystemStore(dbPath string, forceReadOnly bool) (*BBoltSystemStore, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, err
	}

	readOnly := forceReadOnly
	var db *bbolt.DB
	var err error

	if readOnly {
		db, err = bbolt.Open(dbPath, 0666, &bbolt.Options{ReadOnly: true, Timeout: 1 * time.Second})
	} else {
		db, err = bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			if err == bbolt.ErrTimeout {
				// PKS-026 Fallback: Another CLI instance holds the lock.
				readOnly = true
				db, err = bbolt.Open(dbPath, 0666, &bbolt.Options{ReadOnly: true, Timeout: 1 * time.Second})
			}
		}
	}

	if err != nil {
		return nil, err
	}

	if !readOnly {
		// Initialize Logical Buckets
		err = db.Update(func(tx *bbolt.Tx) error {
			buckets := [][]byte{bucketSettings, bucketAuth, bucketTelemetry, bucketMetadata}
			for _, b := range buckets {
				if _, err := tx.CreateBucketIfNotExists(b); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return &BBoltSystemStore{db: db, readOnly: readOnly}, nil
}

func (b *BBoltSystemStore) GetSetting(ctx context.Context, key string) (string, error) {
	var val string
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketSettings)
		if bucket == nil {
			return ports.ErrNotFound
		}
		v := bucket.Get([]byte(key))
		if v == nil {
			return ports.ErrNotFound
		}
		val = string(v)
		return nil
	})
	return val, err
}

func (b *BBoltSystemStore) SetSetting(ctx context.Context, key string, value string) error {
	if b.readOnly {
		return ports.ErrDatabaseLocked
	}
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketSettings)
		if bucket == nil {
			return ports.ErrNotFound
		}
		return bucket.Put([]byte(key), []byte(value))
	})
}

func (b *BBoltSystemStore) SaveAuthToken(ctx context.Context, token string) error {
	if b.readOnly {
		return ports.ErrDatabaseLocked
	}
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketAuth)
		if bucket == nil {
			return ports.ErrNotFound
		}
		return bucket.Put([]byte("jwt_token"), []byte(token))
	})
}

func (b *BBoltSystemStore) GetAuthToken(ctx context.Context) (string, error) {
	var val string
	err := b.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketAuth)
		if bucket == nil {
			return ports.ErrNotFound
		}
		v := bucket.Get([]byte("jwt_token"))
		if v == nil {
			return ports.ErrNotFound
		}
		val = string(v)
		return nil
	})
	return val, err
}

func (b *BBoltSystemStore) LogTelemetry(ctx context.Context, event string, data map[string]any) error {
	if b.readOnly {
		// Silently skip in read-only mode to prevent interrupting the developer workflow.
		return nil
	}
	return b.db.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketTelemetry)
		if bucket == nil {
			return ports.ErrNotFound
		}

		payload, err := json.Marshal(data)
		if err != nil {
			return ports.ErrInvalidPayload
		}

		// Key by timestamp for chronologically ordered retrieval
		key := []byte(time.Now().UTC().Format(time.RFC3339Nano) + "_" + event)
		return bucket.Put(key, payload)
	})
}

func (b *BBoltSystemStore) Close(ctx context.Context) error {
	return b.db.Close()
}
