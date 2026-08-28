package flow018

import (
	"path/filepath"
	"testing"

	"campusqa/internal/review"
	"campusqa/internal/store"
)

func TestRegistryHealthAndWorkflowProgress(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "registry.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	registry := NewRegistry()
	service := New(database)
	if !registry.Register("main", service) || registry.Register("main", service) {
		t.Fatal("unexpected registration result")
	}
	if names := registry.Names(); len(names) != 1 || names[0] != "main" {
		t.Fatalf("unexpected names %#v", names)
	}
	workflow, err := service.WorkflowDefinition("create-review-archive")
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range workflow.Steps {
		workflow = CompletedWorkflow(workflow, step)
	}
	if !workflow.Ready {
		t.Fatal("workflow should be ready")
	}
	if len(WorkflowNames()) != 3 {
		t.Fatal("workflow names missing")
	}
	if err := ValidateSubmission(review.Submission{ID: "id", StudentID: "student", Question: "valid question", Answer: "ok"}); err != nil {
		t.Fatal(err)
	}
	command, err := ParseCommand("search exam schedule")
	if err != nil || command.Name != "search" || command.Record != "exam" || command.Value != "schedule" {
		t.Fatalf("unexpected command %#v %v", command, err)
	}
}
