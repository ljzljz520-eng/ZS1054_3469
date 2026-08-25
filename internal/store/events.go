package store

import (
	"fmt"
	"strconv"

	"campusqa/internal/model"
	"go.etcd.io/bbolt"
)

func (s *Store) AppendAudit(event model.AuditEvent) error {
	data, err := event.Marshal()
	if err != nil {
		return err
	}
	key := []byte(event.ID)
	if len(key) == 0 {
		return fmt.Errorf("audit id is required")
	}
	return s.withUpdate(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketAudits).Put(key, data)
	})
}

func (s *Store) ListAudits(recordID string) ([]model.AuditEvent, error) {
	result := make([]model.AuditEvent, 0)
	err := s.withView(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketAudits)
		for _, key := range sortedKeys(bucket) {
			event, err := model.UnmarshalAudit(bucket.Get(key))
			if err != nil {
				return err
			}
			if event.RecordID == recordID {
				result = append(result, event)
			}
		}
		return nil
	})
	return result, err
}

func (s *Store) NextAuditSequence(recordID string) (int, error) {
	events, err := s.ListAudits(recordID)
	if err != nil {
		return 0, err
	}
	sequence := 0
	for _, event := range events {
		if event.Sequence > sequence {
			sequence = event.Sequence
		}
	}
	return sequence + 1, nil
}

func (s *Store) PutWorkflow(workflow model.Workflow) error {
	if err := model.ValidateWorkflow(workflow); err != nil {
		return err
	}
	data, err := workflow.Marshal()
	if err != nil {
		return err
	}
	return s.withUpdate(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketWorkflows).Put([]byte(workflow.ID), data)
	})
}

func (s *Store) GetWorkflow(id string) (model.Workflow, error) {
	var workflow model.Workflow
	err := s.withView(func(tx *bbolt.Tx) error {
		data := tx.Bucket(bucketWorkflows).Get([]byte(id))
		if data == nil {
			return ErrNotFound
		}
		var err error
		workflow, err = model.UnmarshalWorkflow(data)
		return err
	})
	return workflow, err
}

func (s *Store) PutAttachment(attachment model.Attachment) error {
	if err := model.ValidateAttachment(attachment); err != nil {
		return err
	}
	data, err := attachment.Marshal()
	if err != nil {
		return err
	}
	return s.withUpdate(func(tx *bbolt.Tx) error {
		return tx.Bucket(bucketAttachments).Put([]byte(attachment.ID), data)
	})
}

func (s *Store) ListAttachments(recordID string) ([]model.Attachment, error) {
	result := make([]model.Attachment, 0)
	err := s.withView(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketAttachments)
		for _, key := range sortedKeys(bucket) {
			attachment, err := model.UnmarshalAttachment(bucket.Get(key))
			if err != nil {
				return err
			}
			if attachment.RecordID == recordID {
				result = append(result, attachment)
			}
		}
		return nil
	})
	return result, err
}

func AuditID(recordID string, sequence int) string {
	return recordID + ":audit:" + strconv.Itoa(sequence)
}
