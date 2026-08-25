package model

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMissingID       = errors.New("record id is required")
	ErrMissingStudent  = errors.New("student id is required")
	ErrMissingQuestion = errors.New("question is required")
	ErrMissingAnswer   = errors.New("answer is required")
	ErrMissingCategory = errors.New("category is required")
	ErrBadVersion      = errors.New("version must be positive")
)

func (r Record) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return ErrMissingID
	}
	if strings.TrimSpace(r.StudentID) == "" {
		return ErrMissingStudent
	}
	if strings.TrimSpace(r.Question) == "" {
		return ErrMissingQuestion
	}
	if strings.TrimSpace(r.Answer) == "" {
		return ErrMissingAnswer
	}
	if strings.TrimSpace(r.Category) == "" {
		return ErrMissingCategory
	}
	if r.Version < 1 {
		return ErrBadVersion
	}
	if !IsKnownStatus(r.Status) {
		return fmt.Errorf("unknown status %q", r.Status)
	}
	if len(r.Question) > 500 {
		return errors.New("question exceeds 500 characters")
	}
	if len(r.Answer) > 4000 {
		return errors.New("answer exceeds 4000 characters")
	}
	return nil
}

func (q Query) Validate() error {
	if q.Limit < 0 {
		return errors.New("limit cannot be negative")
	}
	if q.Limit > 100 {
		return errors.New("limit cannot exceed 100")
	}
	if q.Status != "" && !IsKnownStatus(q.Status) {
		return fmt.Errorf("unknown query status %q", q.Status)
	}
	return nil
}

func (d ReviewDecision) Validate() error {
	if strings.TrimSpace(d.RecordID) == "" {
		return ErrMissingID
	}
	if strings.TrimSpace(d.Reviewer) == "" {
		return errors.New("reviewer is required")
	}
	if !d.Approve && strings.TrimSpace(d.Reason) == "" {
		return errors.New("rejection reason is required")
	}
	return nil
}

func IsKnownStatus(status RecordStatus) bool {
	switch status {
	case StatusDraft, StatusSubmitted, StatusApproved, StatusPublished, StatusArchived, StatusRejected:
		return true
	default:
		return false
	}
}

func ValidateAttachment(a Attachment) error {
	if strings.TrimSpace(a.ID) == "" || strings.TrimSpace(a.RecordID) == "" {
		return errors.New("attachment identity is required")
	}
	if strings.TrimSpace(a.Name) == "" {
		return errors.New("attachment name is required")
	}
	if a.Size < 0 || a.Size != len(a.Content) {
		return errors.New("attachment size does not match content")
	}
	if strings.TrimSpace(a.Checksum) == "" {
		return errors.New("attachment checksum is required")
	}
	return nil
}

func ValidateWorkflow(w Workflow) error {
	if w.ID == "" || w.RecordID == "" || w.Name == "" {
		return errors.New("workflow identity is required")
	}
	if len(w.Steps) == 0 {
		return errors.New("workflow needs steps")
	}
	if len(w.Completed) > len(w.Steps) {
		return errors.New("workflow completed steps exceed plan")
	}
	return nil
}
