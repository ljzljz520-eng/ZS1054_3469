package store

import (
	"fmt"

	"campusqa/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) PutRecord(record model.Record) error {
	if err := record.Validate(); err != nil {
		return err
	}
	data, err := record.Marshal()
	if err != nil {
		return fmt.Errorf("encode record: %w", err)
	}
	return s.withUpdate(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketRecords).Put([]byte(record.ID), data)
	})
}

func (s *Store) GetRecord(id string) (model.Record, error) {
	var record model.Record
	err := s.withView(func(tx *bbolt.Tx) error {
		value := tx.Bucket(bucketRecords).Get([]byte(id))
		if value == nil {
			return ErrNotFound
		}
		var err error
		record, err = model.UnmarshalRecord(value)
		return err
	})
	if err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func (s *Store) DeleteRecord(id string) error {
	return s.withUpdate(func(tx *bbolt.Tx) error {
		if tx.Bucket(bucketRecords).Get([]byte(id)) == nil {
			return ErrNotFound
		}
		return tx.Bucket(bucketRecords).Delete([]byte(id))
	})
}

func (s *Store) ListRecords() ([]model.Record, error) {
	result := make([]model.Record, 0)
	err := s.withView(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketRecords)
		for _, key := range sortedKeys(bucket) {
			value := bucket.Get(key)
			record, err := model.UnmarshalRecord(value)
			if err != nil {
				return err
			}
			result = append(result, record)
		}
		return nil
	})
	return result, err
}

func (s *Store) UpdateRecord(id string, update func(*model.Record) error) (model.Record, error) {
	var result model.Record
	err := s.withUpdate(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketRecords)
		data := bucket.Get([]byte(id))
		if data == nil {
			return ErrNotFound
		}
		record, err := model.UnmarshalRecord(data)
		if err != nil {
			return err
		}
		if err := update(&record); err != nil {
			return err
		}
		if err := record.Validate(); err != nil {
			return err
		}
		encoded, err := record.Marshal()
		if err != nil {
			return err
		}
		if err := bucket.Put([]byte(id), encoded); err != nil {
			return err
		}
		result = record
		return nil
	})
	return result, err
}

func (s *Store) CountRecords() (int, error) {
	count := 0
	err := s.withView(func(tx *bbolt.Tx) error {
		count = tx.Bucket(bucketRecords).Stats().KeyN
		return nil
	})
	return count, err
}
