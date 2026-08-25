package flow018

import (
	"path/filepath"
	"testing"

	"campusqa/internal/model"
	"campusqa/internal/review"
	"campusqa/internal/store"
)

func TestWorkflowCreateReviewArchive(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "flow.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	receipt, err := New(database).CreateReviewArchive(review.Submission{ID: "wf-1", StudentID: "s1", Question: "How do I enroll in a course?", Answer: "Use portal"}, "reviewer")
	if err != nil || receipt.Status != model.StatusArchived || receipt.AuditCount != 4 {
		t.Fatalf("workflow failed: %v %#v", err, receipt)
	}
}

func TestWorkflowSearchUpdatePublish(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "flow-update.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := New(database)
	record, err := service.review.Submit(review.Submission{ID: "wf-2", StudentID: "s1", Question: "What is the exam date?", Answer: "old"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.review.Decide(model.ReviewDecision{RecordID: record.ID, Reviewer: "reviewer", Approve: true, Reason: "ok"}); err != nil {
		t.Fatal(err)
	}
	receipt, err := service.SearchUpdatePublish(record.ID, "reviewer", "new")
	if err != nil || receipt.Status != model.StatusPublished {
		t.Fatalf("update workflow failed: %v %#v", err, receipt)
	}
}

func TestWorkflowImportReport(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "flow-import.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	summary, imported, err := New(database).ImportReport("r1|s1|course enrollment|portal|general\nr2|s2|exam date|calendar|general")
	if err != nil || imported.Accepted != 2 || summary.Total != 2 {
		t.Fatalf("import workflow failed: %v %#v %#v", err, imported, summary)
	}
}

func Test1054BusinessRegression(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "regression.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := New(database)
	first, err := service.review.Submit(review.Submission{ID: "same-question", StudentID: "s1", Question: "course exam", Answer: "first"})
	if err != nil {
		t.Fatal(err)
	}
	if first.Category != "examination" {
		t.Fatalf("unexpected initial category %s", first.Category)
	}
	if _, err := service.review.Decide(model.ReviewDecision{RecordID: first.ID, Reviewer: "reviewer", Approve: false, Reason: "clarify"}); err != nil {
		t.Fatal(err)
	}
	second, err := service.review.Submit(review.Submission{ID: first.ID, StudentID: "s1", Question: "course exam", Answer: "second"})
	if err != nil {
		t.Fatal(err)
	}
	if second.Category != first.Category {
		t.Fatalf("resubmission category changed from %s to %s", first.Category, second.Category)
	}
}
