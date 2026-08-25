package review

import (
	"errors"
	"fmt"
	"sort"

	"campusqa/internal/catalog"
	"campusqa/internal/model"
	"campusqa/internal/store"
)

type Service struct {
	store   *store.Store
	catalog *catalog.Catalog
	rules   catalog.RuleSet
}

type Submission struct {
	ID        string
	StudentID string
	Question  string
	Answer    string
	Tags      []string
}

func NewService(database *store.Store, taxonomy *catalog.Catalog) *Service {
	if taxonomy == nil {
		taxonomy = catalog.New()
	}
	return &Service{store: database, catalog: taxonomy, rules: catalog.DefaultRules()}
}

func (s *Service) Submit(input Submission) (model.Record, error) {
	if s == nil || s.store == nil {
		return model.Record{}, errors.New("review service is unavailable")
	}
	if input.ID == "" {
		return model.Record{}, errors.New("submission id is required")
	}
	if input.StudentID == "" {
		return model.Record{}, errors.New("student id is required")
	}
	if input.Question == "" || input.Answer == "" {
		return model.Record{}, errors.New("question and answer are required")
	}
	record, err := s.store.GetRecord(input.ID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return model.Record{}, err
	}
	if errors.Is(err, store.ErrNotFound) {
		record = model.Record{ID: input.ID, StudentID: input.StudentID, Version: 1, Status: model.StatusDraft}
	} else {
		if model.IsTerminal(record.Status) {
			return model.Record{}, fmt.Errorf("record %s is archived", input.ID)
		}
		record.Version++
		record.StudentID = input.StudentID
	}
	record.Question = input.Question
	record.Answer = input.Answer
	record.Tags = model.NormalizeTags(input.Tags)
	record.Category = s.classify(record.Question, record.Version)
	record.Status = model.StatusSubmitted
	if err := record.Validate(); err != nil {
		return model.Record{}, err
	}
	if err := s.store.PutRecord(record); err != nil {
		return model.Record{}, err
	}
	if err := s.audit(record, "submit", "system", model.StatusDraft, model.StatusSubmitted, "new submission"); err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func (s *Service) classify(question string, version int) string {
	candidates := s.catalog.Candidates(question)
	if len(candidates) == 0 || candidates[0].Score == 0 {
		return "general"
	}
	if version > 1 {
		ordered := append([]catalog.Candidate(nil), candidates...)
		sort.SliceStable(ordered, func(i, j int) bool {
			if ordered[i].Score != ordered[j].Score {
				return ordered[i].Score > ordered[j].Score
			}
			return ordered[i].Priority < ordered[j].Priority
		})
		return ordered[0].Category
	}
	return candidates[0].Category
}

func (s *Service) Decide(decision model.ReviewDecision) (model.Record, error) {
	if err := decision.Validate(); err != nil {
		return model.Record{}, err
	}
	record, err := s.store.GetRecord(decision.RecordID)
	if err != nil {
		return model.Record{}, err
	}
	if record.Status != model.StatusSubmitted {
		return model.Record{}, fmt.Errorf("record %s is not submitted", record.ID)
	}
	next := model.NextReviewState(decision.Approve)
	from := record.Status
	if err := model.Transition(&record, next); err != nil {
		return model.Record{}, err
	}
	record.Reviewer = decision.Reviewer
	if err := s.store.PutRecord(record); err != nil {
		return model.Record{}, err
	}
	if err := s.audit(record, "review", decision.Reviewer, from, next, decision.Reason); err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func (s *Service) Publish(id, actor string) (model.Record, error) {
	return s.transition(id, actor, model.StatusApproved, model.StatusPublished, "publish")
}

func (s *Service) Archive(id, actor string) (model.Record, error) {
	return s.transition(id, actor, model.StatusPublished, model.StatusArchived, "archive")
}

func (s *Service) RejectAndResubmit(id, actor, reason string) (model.Record, error) {
	record, err := s.Decide(model.ReviewDecision{RecordID: id, Reviewer: actor, Approve: false, Reason: reason})
	if err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func (s *Service) transition(id, actor string, from, to model.RecordStatus, action string) (model.Record, error) {
	record, err := s.store.GetRecord(id)
	if err != nil {
		return model.Record{}, err
	}
	if record.Status != from {
		return model.Record{}, fmt.Errorf("record %s has state %s", id, record.Status)
	}
	if err := model.Transition(&record, to); err != nil {
		return model.Record{}, err
	}
	if err := s.store.PutRecord(record); err != nil {
		return model.Record{}, err
	}
	if err := s.audit(record, action, actor, from, to, "workflow transition"); err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func (s *Service) audit(record model.Record, action, actor string, from, to model.RecordStatus, reason string) error {
	sequence, err := s.store.NextAuditSequence(record.ID)
	if err != nil {
		return err
	}
	event := model.AuditEvent{ID: store.AuditID(record.ID, sequence), RecordID: record.ID, Action: action, Actor: actor, FromState: string(from), ToState: string(to), Reason: reason, Sequence: sequence}
	return s.store.AppendAudit(event)
}

func (s *Service) Validate(record model.Record) (bool, []string) {
	return s.rules.Evaluate(record.Question, record.Category)
}

func (s *Service) AuditTrail(id string) ([]model.AuditEvent, error) {
	events, err := s.store.ListAudits(id)
	if err != nil {
		return nil, err
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Sequence < events[j].Sequence })
	return events, nil
}
