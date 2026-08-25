package store

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"

	"go.etcd.io/bbolt"
)

var (
	ErrNotFound = errors.New("entity not found")
	ErrClosed   = errors.New("store is closed")
)

var bucketRecords = []byte("records")
var bucketAudits = []byte("audits")
var bucketWorkflows = []byte("workflows")
var bucketAttachments = []byte("attachments")
var bucketMeta = []byte("meta")

type Store struct {
	db   *bbolt.DB
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	db, err := bbolt.Open(filepath.Clean(path), 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	store := &Store{db: db, path: path}
	if err := store.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize() error {
	if s == nil || s.db == nil {
		return ErrClosed
	}
	return s.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{bucketRecords, bucketAudits, bucketWorkflows, bucketAttachments, bucketMeta} {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}
		meta := tx.Bucket(bucketMeta)
		if meta.Get([]byte("schema")) == nil {
			return meta.Put([]byte("schema"), []byte("1"))
		}
		return nil
	})
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func (s *Store) withView(fn func(*bbolt.Tx) error) error {
	if s == nil || s.db == nil {
		return ErrClosed
	}
	return s.db.View(fn)
}

func (s *Store) withUpdate(fn func(*bbolt.Tx) error) error {
	if s == nil || s.db == nil {
		return ErrClosed
	}
	return s.db.Update(fn)
}

func sortedKeys(bucket *bbolt.Bucket) [][]byte {
	keys := make([][]byte, 0)
	_ = bucket.ForEach(func(key, value []byte) error {
		if value != nil {
			keys = append(keys, append([]byte(nil), key...))
		}
		return nil
	})
	sort.Slice(keys, func(i, j int) bool { return string(keys[i]) < string(keys[j]) })
	return keys
}
