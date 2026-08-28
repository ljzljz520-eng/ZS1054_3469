package flow018

import (
	"errors"
	"fmt"
	"strings"

	"campusqa/internal/catalog"
	"campusqa/internal/model"
	"campusqa/internal/report"
	"campusqa/internal/review"
	"campusqa/internal/search"
	"campusqa/internal/store"
)

type Service struct {
	store  *store.Store
	review *review.Service
	search *search.Service
	index  *search.Index
}

type WorkflowReceipt struct {
	WorkflowID string
	RecordID   string
	Status     model.RecordStatus
	Category   string
	AuditCount int
}

func New(database *store.Store) *Service {
	return &Service{store: database, review: review.NewService(database, catalog.New()), search: search.NewService(database), index: search.NewIndex()}
}

func (s *Service) CreateReviewArchive(input review.Submission, reviewer string) (WorkflowReceipt, error) {
	if strings.TrimSpace(reviewer) == "" {
		return WorkflowReceipt{}, errors.New("reviewer is required")
	}
	record, err := s.review.Submit(input)
	if err != nil {
		return WorkflowReceipt{}, err
	}
	approved, issues := s.review.Validate(record)
	if !approved {
		return WorkflowReceipt{}, fmt.Errorf("validation failed: %s", strings.Join(issues, ", "))
	}
	record, err = s.review.Decide(model.ReviewDecision{RecordID: record.ID, Reviewer: reviewer, Approve: true, Reason: "validated"})
	if err != nil {
		return WorkflowReceipt{}, err
	}
	record, err = s.review.Publish(record.ID, reviewer)
	if err != nil {
		return WorkflowReceipt{}, err
	}
	record, err = s.review.Archive(record.ID, reviewer)
	if err != nil {
		return WorkflowReceipt{}, err
	}
	s.index.Add(record)
	audits, err := s.review.AuditTrail(record.ID)
	if err != nil {
		return WorkflowReceipt{}, err
	}
	return WorkflowReceipt{WorkflowID: "create-review-archive", RecordID: record.ID, Status: record.Status, Category: record.Category, AuditCount: len(audits)}, nil
}

func (s *Service) SearchUpdatePublish(id, actor, replacement string) (WorkflowReceipt, error) {
	if replacement == "" {
		return WorkflowReceipt{}, errors.New("replacement answer is required")
	}
	record, err := s.store.GetRecord(id)
	if err != nil {
		return WorkflowReceipt{}, err
	}
	if record.Status == model.StatusArchived {
		return WorkflowReceipt{}, errors.New("archived record cannot be updated")
	}
	results, err := s.search.Search(model.Query{Text: record.Question, Limit: 10})
	if err != nil {
		return WorkflowReceipt{}, err
	}
	selected := false
	for _, result := range results {
		if result.Record.ID == id {
			selected = true
			break
		}
	}
	if !selected {
		return WorkflowReceipt{}, errors.New("record is not searchable")
	}
	if record.Status == model.StatusPublished {
		return WorkflowReceipt{}, errors.New("published record must be archived before update")
	}
	record, err = s.store.UpdateRecord(id, func(target *model.Record) error {
		target.Answer = replacement
		target.Version++
		return nil
	})
	if err != nil {
		return WorkflowReceipt{}, err
	}
	if record.Status == model.StatusRejected {
		record.Status = model.StatusSubmitted
		if err := s.store.PutRecord(record); err != nil {
			return WorkflowReceipt{}, err
		}
	}
	if record.Status == model.StatusApproved {
		record, err = s.review.Publish(id, actor)
		if err != nil {
			return WorkflowReceipt{}, err
		}
	}
	audits, err := s.review.AuditTrail(id)
	if err != nil {
		return WorkflowReceipt{}, err
	}
	s.index.Add(record)
	return WorkflowReceipt{WorkflowID: "search-update-publish", RecordID: id, Status: record.Status, Category: record.Category, AuditCount: len(audits)}, nil
}

func (s *Service) ImportReport(input string) (report.Summary, report.ImportSummary, error) {
	records, summary, err := report.ParseLines(input)
	if err != nil {
		return report.Summary{}, summary, err
	}
	for _, record := range records {
		if record.Category == "general" {
			record.Category = catalog.New().Classify(record.Question)
		}
		if err := s.store.PutRecord(record); err != nil {
			return report.Summary{}, summary, err
		}
		s.index.Add(record)
	}
	all, err := s.store.ListRecords()
	if err != nil {
		return report.Summary{}, summary, err
	}
	return report.Build(all), summary, nil
}

func (s *Service) AddAttachment(recordID, id, name, mediaType, content string) (model.Attachment, error) {
	attachment := model.Attachment{ID: id, RecordID: recordID, Name: name, MediaType: mediaType, Content: content, Size: len(content), Checksum: checksum(content)}
	if err := s.store.PutAttachment(attachment); err != nil {
		return model.Attachment{}, err
	}
	return attachment, nil
}

func checksum(value string) string {
	total := 0
	for index, r := range value {
		total += int(r) * (index + 1)
	}
	return fmt.Sprintf("sum-%d", total)
}

func (s *Service) Records() ([]model.Record, error) {
	return s.store.ListRecords()
}

func (s *Service) Find(query model.Query) ([]search.Result, error) {
	return s.search.Search(query)
}

func (s *Service) WorkflowDefinition(name string) (model.Workflow, error) {
	definitions := map[string]model.Workflow{
		"create-review-archive": {ID: "wf-create", Name: "create-review-archive", Stage: "archive", Owner: "registrar", Steps: []string{"create", "review", "confirm", "archive"}, Ready: true},
		"search-update-publish": {ID: "wf-update", Name: "search-update-publish", Stage: "publish", Owner: "registrar", Steps: []string{"search", "select", "update", "publish"}, Ready: true},
		"import-report":         {ID: "wf-import", Name: "import-report", Stage: "report", Owner: "registrar", Steps: []string{"import", "validate", "persist", "report"}, Ready: true},
	}
	workflow, ok := definitions[name]
	if !ok {
		return model.Workflow{}, fmt.Errorf("unknown workflow %s", name)
	}
	return workflow, nil
}
