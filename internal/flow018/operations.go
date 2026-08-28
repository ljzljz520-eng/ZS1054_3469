package flow018

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"campusqa/internal/model"
	"campusqa/internal/report"
	"campusqa/internal/review"
)

type IntakeResult struct {
	Receipts []WorkflowReceipt
	Issues   []string
}

type Dashboard struct {
	Total      int
	Pending    int
	Published  int
	Archived   int
	Categories map[string]int
	Reviewers  map[string]int
}

type Command struct {
	Name   string
	Record string
	Value  string
}

func (s *Service) IntakeBatch(inputs []review.Submission, reviewer string) IntakeResult {
	result := IntakeResult{Receipts: make([]WorkflowReceipt, 0), Issues: make([]string, 0)}
	for _, input := range inputs {
		receipt, err := s.CreateReviewArchive(input, reviewer)
		if err != nil {
			result.Issues = append(result.Issues, input.ID+": "+err.Error())
			continue
		}
		result.Receipts = append(result.Receipts, receipt)
	}
	return result
}

func (s *Service) ReviewQueue() ([]model.Record, error) {
	records, err := s.store.ListRecords()
	if err != nil {
		return nil, err
	}
	queue := review.NewQueue(records)
	return queue.Pending(), nil
}

func (s *Service) ArchiveEligible(actor string) ([]WorkflowReceipt, error) {
	if strings.TrimSpace(actor) == "" {
		return nil, errors.New("archive actor is required")
	}
	records, err := s.store.ListRecords()
	if err != nil {
		return nil, err
	}
	result := make([]WorkflowReceipt, 0)
	for _, record := range records {
		if record.Status != model.StatusPublished {
			continue
		}
		archived, err := s.review.Archive(record.ID, actor)
		if err != nil {
			return nil, err
		}
		audits, err := s.review.AuditTrail(record.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, WorkflowReceipt{WorkflowID: "archive-batch", RecordID: archived.ID, Status: archived.Status, Category: archived.Category, AuditCount: len(audits)})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].RecordID < result[j].RecordID })
	return result, nil
}

func (s *Service) Dashboard() (Dashboard, error) {
	records, err := s.store.ListRecords()
	if err != nil {
		return Dashboard{}, err
	}
	dashboard := Dashboard{Categories: make(map[string]int), Reviewers: make(map[string]int)}
	for _, record := range records {
		dashboard.Total++
		dashboard.Categories[record.Category]++
		if record.Reviewer != "" {
			dashboard.Reviewers[record.Reviewer]++
		}
		switch record.Status {
		case model.StatusSubmitted:
			dashboard.Pending++
		case model.StatusPublished:
			dashboard.Published++
		case model.StatusArchived:
			dashboard.Archived++
		}
	}
	return dashboard, nil
}

func ParseCommand(line string) (Command, error) {
	parts := strings.Fields(strings.TrimSpace(line))
	if len(parts) == 0 {
		return Command{}, errors.New("command is empty")
	}
	command := Command{Name: strings.ToLower(parts[0])}
	if len(parts) > 1 {
		command.Record = parts[1]
	}
	if len(parts) > 2 {
		command.Value = strings.Join(parts[2:], " ")
	}
	switch command.Name {
	case "search", "archive", "publish", "review", "import":
		return command, nil
	default:
		return Command{}, fmt.Errorf("unknown command %s", command.Name)
	}
}

func (s *Service) ExecuteCommand(command Command, actor string) (string, error) {
	if command.Name == "" {
		return "", errors.New("command name is required")
	}
	switch command.Name {
	case "search":
		results, err := s.Find(model.Query{Text: command.Value, Limit: 20})
		if err != nil {
			return "", err
		}
		ids := make([]string, 0, len(results))
		for _, result := range results {
			ids = append(ids, result.Record.ID)
		}
		return strings.Join(ids, ","), nil
	case "archive":
		record, err := s.review.Archive(command.Record, actor)
		if err != nil {
			return "", err
		}
		return string(record.Status), nil
	case "publish":
		record, err := s.review.Publish(command.Record, actor)
		if err != nil {
			return "", err
		}
		return string(record.Status), nil
	case "review":
		approve := strings.EqualFold(command.Value, "approve")
		record, err := s.review.Decide(model.ReviewDecision{RecordID: command.Record, Reviewer: actor, Approve: approve, Reason: command.Value})
		if err != nil {
			return "", err
		}
		return string(record.Status), nil
	case "import":
		summary, imported, err := s.ImportReport(command.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("accepted=%d total=%d", imported.Accepted, summary.Total), nil
	default:
		return "", fmt.Errorf("unsupported command %s", command.Name)
	}
}

func (s *Service) Export(filter report.Filter) (string, error) {
	records, err := s.Records()
	if err != nil {
		return "", err
	}
	filtered := report.ApplyFilter(records, filter)
	var builder strings.Builder
	if err := report.WriteCSV(&builder, filtered); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func ValidateSubmission(input review.Submission) error {
	if strings.TrimSpace(input.ID) == "" {
		return errors.New("submission id is required")
	}
	if strings.TrimSpace(input.StudentID) == "" {
		return errors.New("student id is required")
	}
	if len(strings.TrimSpace(input.Question)) < 5 {
		return errors.New("question is too short")
	}
	if len(strings.TrimSpace(input.Answer)) < 2 {
		return errors.New("answer is too short")
	}
	return nil
}

func WorkflowNames() []string {
	names := []string{"create-review-archive", "search-update-publish", "import-report"}
	sort.Strings(names)
	return names
}

func (s *Service) Reopen(id, actor string) (model.Record, error) {
	if strings.TrimSpace(actor) == "" {
		return model.Record{}, errors.New("reopen actor is required")
	}
	record, err := s.store.GetRecord(id)
	if err != nil {
		return model.Record{}, err
	}
	if record.Status != model.StatusRejected {
		return model.Record{}, fmt.Errorf("record %s is not rejected", id)
	}
	record.Status = model.StatusDraft
	record.Version++
	if err := s.store.PutRecord(record); err != nil {
		return model.Record{}, err
	}
	sequence, err := s.store.NextAuditSequence(id)
	if err != nil {
		return model.Record{}, err
	}
	event := model.AuditEvent{ID: storeAuditID(id, sequence), RecordID: id, Action: "reopen", Actor: actor, FromState: string(model.StatusRejected), ToState: string(model.StatusDraft), Reason: "requested changes", Sequence: sequence}
	if err := s.store.AppendAudit(event); err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func storeAuditID(recordID string, sequence int) string {
	return fmt.Sprintf("%s:audit:%d", recordID, sequence)
}

func (s *Service) ExportSummary() (string, error) {
	dashboard, err := s.Dashboard()
	if err != nil {
		return "", err
	}
	categoryNames := make([]string, 0, len(dashboard.Categories))
	for category := range dashboard.Categories {
		categoryNames = append(categoryNames, category)
	}
	sort.Strings(categoryNames)
	categoryValues := make([]string, 0, len(categoryNames))
	for _, category := range categoryNames {
		categoryValues = append(categoryValues, category+"="+fmt.Sprint(dashboard.Categories[category]))
	}
	return fmt.Sprintf("total=%d pending=%d published=%d archived=%d categories=%s", dashboard.Total, dashboard.Pending, dashboard.Published, dashboard.Archived, strings.Join(categoryValues, ",")), nil
}

func (s *Service) QueueSnapshot() ([]model.Record, error) {
	queue, err := s.ReviewQueue()
	if err != nil {
		return nil, err
	}
	return model.SortRecords(queue, false), nil
}
