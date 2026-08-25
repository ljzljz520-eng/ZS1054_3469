package model

import "testing"

func TestRecordValidationAndTransitions(t *testing.T) {
	record := Record{ID: "r1", StudentID: "s1", Question: "question", Answer: "answer", Category: "general", Status: StatusDraft, Version: 1}
	if err := record.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, next := range []RecordStatus{StatusSubmitted, StatusApproved, StatusPublished, StatusArchived} {
		if err := Transition(&record, next); err != nil {
			t.Fatal(err)
		}
	}
	if !IsTerminal(record.Status) {
		t.Fatalf("expected terminal status, got %s", record.Status)
	}
}

func TestNormalizationAndAttachment(t *testing.T) {
	if got := NormalizeText("  Course   ENROLL  "); got != "course enroll" {
		t.Fatalf("unexpected normalization %q", got)
	}
	attachment := Attachment{ID: "a1", RecordID: "r1", Name: "guide.txt", Content: "abc", Size: 3, Checksum: "sum"}
	if err := ValidateAttachment(attachment); err != nil {
		t.Fatal(err)
	}
	changed := recordForCompare()
	changes := CompareRecords(changed, Record{ID: "r1", StudentID: "s1", Question: "new", Answer: "answer", Category: "general", Status: StatusDraft, Version: 1})
	if len(changes) != 1 || RecordFingerprint(changed) == "" {
		t.Fatalf("unexpected comparison %#v", changes)
	}
	if sorted := SortRecords([]Record{changed}, true); len(sorted) != 1 {
		t.Fatal("record sort failed")
	}
}

func recordForCompare() Record {
	return Record{ID: "r1", StudentID: "s1", Question: "question", Answer: "answer", Category: "general", Status: StatusDraft, Version: 1}
}
