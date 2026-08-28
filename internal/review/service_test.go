package review

import (
	"path/filepath"
	"testing"

	"campusqa/internal/model"
	"campusqa/internal/store"
)

func TestSubmitReviewAndAudit(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "review.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := NewService(database, nil)
	record, err := service.Submit(Submission{ID: "review-1", StudentID: "student", Question: "How do I enroll in a course?", Answer: "Use portal"})
	if err != nil || record.Status != model.StatusSubmitted {
		t.Fatalf("submit failed: %v %#v", err, record)
	}
	approved, err := service.Decide(model.ReviewDecision{RecordID: record.ID, Reviewer: "staff", Approve: true, Reason: "clear"})
	if err != nil || approved.Status != model.StatusApproved {
		t.Fatalf("decision failed: %v %#v", err, approved)
	}
	events, err := service.AuditTrail(record.ID)
	if err != nil || len(events) != 2 {
		t.Fatalf("audit failed: %v %#v", err, events)
	}
	policy := EvaluatePolicy(record, DefaultPolicies())
	if QueuePriority(record, policy) <= 0 || len(PolicyNames(DefaultPolicies())) != 3 {
		t.Fatal("policy evaluation failed")
	}
}
