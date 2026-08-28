package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RecordStatus string

const (
	StatusDraft     RecordStatus = "draft"
	StatusSubmitted RecordStatus = "submitted"
	StatusApproved  RecordStatus = "approved"
	StatusPublished RecordStatus = "published"
	StatusArchived  RecordStatus = "archived"
	StatusRejected  RecordStatus = "rejected"
)

type Record struct {
	ID          string       `json:"id"`
	StudentID   string       `json:"student_id"`
	Question    string       `json:"question"`
	Category    string       `json:"category"`
	Answer      string       `json:"answer"`
	Status      RecordStatus `json:"status"`
	Version     int          `json:"version"`
	SubmittedAt int64        `json:"submitted_at"`
	UpdatedAt   int64        `json:"updated_at"`
	Reviewer    string       `json:"reviewer"`
	Tags        []string     `json:"tags"`
}

type AuditEvent struct {
	ID        string `json:"id"`
	RecordID  string `json:"record_id"`
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	FromState string `json:"from_state"`
	ToState   string `json:"to_state"`
	Reason    string `json:"reason"`
	Sequence  int    `json:"sequence"`
}

type Workflow struct {
	ID        string   `json:"id"`
	RecordID  string   `json:"record_id"`
	Name      string   `json:"name"`
	Stage     string   `json:"stage"`
	Owner     string   `json:"owner"`
	Steps     []string `json:"steps"`
	Completed []string `json:"completed"`
	Ready     bool     `json:"ready"`
}

type Attachment struct {
	ID        string `json:"id"`
	RecordID  string `json:"record_id"`
	Name      string `json:"name"`
	MediaType string `json:"media_type"`
	Content   string `json:"content"`
	Checksum  string `json:"checksum"`
	Size      int    `json:"size"`
}

type Query struct {
	Text      string
	Category  string
	StudentID string
	Status    RecordStatus
	Limit     int
}

type ReviewDecision struct {
	RecordID string
	Reviewer string
	Approve  bool
	Reason   string
}

func (r Record) Clone() Record {
	copyTags := append([]string(nil), r.Tags...)
	r.Tags = copyTags
	return r
}

func (r Record) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalRecord(data []byte) (Record, error) {
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return Record{}, fmt.Errorf("decode record: %w", err)
	}
	return record, nil
}

func (e AuditEvent) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func UnmarshalAudit(data []byte) (AuditEvent, error) {
	var event AuditEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return AuditEvent{}, fmt.Errorf("decode audit: %w", err)
	}
	return event, nil
}

func (w Workflow) Marshal() ([]byte, error) {
	return json.Marshal(w)
}

func UnmarshalWorkflow(data []byte) (Workflow, error) {
	var workflow Workflow
	if err := json.Unmarshal(data, &workflow); err != nil {
		return Workflow{}, fmt.Errorf("decode workflow: %w", err)
	}
	return workflow, nil
}

func (a Attachment) Marshal() ([]byte, error) {
	return json.Marshal(a)
}

func UnmarshalAttachment(data []byte) (Attachment, error) {
	var attachment Attachment
	if err := json.Unmarshal(data, &attachment); err != nil {
		return Attachment{}, fmt.Errorf("decode attachment: %w", err)
	}
	return attachment, nil
}

func NormalizeText(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(value)), " ")
}

func NormalizeTags(tags []string) []string {
	seen := make(map[string]bool, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := NormalizeText(tag)
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		result = append(result, normalized)
	}
	return result
}
