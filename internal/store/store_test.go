package store

import (
	"path/filepath"
	"testing"

	"campusqa/internal/model"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "campusqa.db")
	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	record := model.Record{ID: "persist-1", StudentID: "student-1", Question: "where is the exam?", Answer: "calendar", Category: "examination", Status: model.StatusSubmitted, Version: 1}
	if err := first.PutRecord(record); err != nil {
		t.Fatal(err)
	}
	if err := first.AppendAudit(model.AuditEvent{ID: "persist-1:audit:1", RecordID: "persist-1", Action: "submit", ToState: "submitted", Sequence: 1}); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	got, err := second.GetRecord("persist-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Question != record.Question || got.Category != record.Category {
		t.Fatalf("unexpected record after reopen: %#v", got)
	}
	events, err := second.ListAudits("persist-1")
	if err != nil || len(events) != 1 {
		t.Fatalf("unexpected audit after reopen: %v %#v", err, events)
	}
}

func TestStoreUpdateAndLists(t *testing.T) {
	database, err := Open(filepath.Join(t.TempDir(), "store.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	record := model.Record{ID: "r1", StudentID: "s1", Question: "q", Answer: "a", Category: "general", Status: model.StatusDraft, Version: 1}
	if err := database.PutRecord(record); err != nil {
		t.Fatal(err)
	}
	updated, err := database.UpdateRecord("r1", func(target *model.Record) error { target.Answer = "changed"; return nil })
	if err != nil || updated.Answer != "changed" {
		t.Fatalf("update failed: %v %#v", err, updated)
	}
	count, err := database.CountRecords()
	if err != nil || count != 1 {
		t.Fatalf("count failed: %v %d", err, count)
	}
	if found, err := database.HasRecord("r1"); err != nil || !found {
		t.Fatalf("record existence failed: %v %v", err, found)
	}
	batch, err := database.PutRecords([]model.Record{{ID: "r2", StudentID: "s2", Question: "q2", Answer: "a2", Category: "general", Status: model.StatusDraft, Version: 1}})
	if err != nil || batch.Inserted != 1 {
		t.Fatalf("batch put failed: %v %#v", err, batch)
	}
	if snapshot, err := database.Snapshot(); err != nil || len(snapshot) != 2 {
		t.Fatalf("snapshot failed: %v %#v", err, snapshot)
	}
	replaced, err := database.ReplaceRecords([]model.Record{{ID: "r3", StudentID: "s3", Question: "q3", Answer: "a3", Category: "general", Status: model.StatusDraft, Version: 1}})
	if err != nil || replaced.Inserted != 1 {
		t.Fatalf("replace failed: %v %#v", err, replaced)
	}
}
