package store

import (
	"fmt"
	"sort"

	"campusqa/internal/model"
	"go.etcd.io/bbolt"
)

type BatchResult struct {
	Inserted int
	Updated  int
	Rejected int
	Issues   []string
}

func (s *Store) PutRecords(records []model.Record) (BatchResult, error) {
	result := BatchResult{Issues: make([]string, 0)}
	err := s.withUpdate(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketRecords)
		for _, record := range records {
			if err := record.Validate(); err != nil {
				result.Rejected++
				result.Issues = append(result.Issues, fmt.Sprintf("%s: %v", record.ID, err))
				continue
			}
			data, err := record.Marshal()
			if err != nil {
				return err
			}
			if bucket.Get([]byte(record.ID)) == nil {
				result.Inserted++
			} else {
				result.Updated++
			}
			if err := bucket.Put([]byte(record.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}

func (s *Store) Snapshot() ([]model.Record, error) {
	records, err := s.ListRecords()
	if err != nil {
		return nil, err
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Category != records[j].Category {
			return records[i].Category < records[j].Category
		}
		return records[i].ID < records[j].ID
	})
	return records, nil
}

func (s *Store) ReplaceRecords(records []model.Record) (BatchResult, error) {
	result := BatchResult{Issues: make([]string, 0)}
	err := s.withUpdate(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketRecords)
		for _, key := range sortedKeys(bucket) {
			if err := bucket.Delete(key); err != nil {
				return err
			}
		}
		for _, record := range records {
			if err := record.Validate(); err != nil {
				result.Rejected++
				result.Issues = append(result.Issues, fmt.Sprintf("%s: %v", record.ID, err))
				continue
			}
			data, err := record.Marshal()
			if err != nil {
				return err
			}
			if err := bucket.Put([]byte(record.ID), data); err != nil {
				return err
			}
			result.Inserted++
		}
		return nil
	})
	return result, err
}

func (s *Store) HasRecord(id string) (bool, error) {
	found := false
	err := s.withView(func(tx *bbolt.Tx) error {
		found = tx.Bucket(bucketRecords).Get([]byte(id)) != nil
		return nil
	})
	return found, err
}
