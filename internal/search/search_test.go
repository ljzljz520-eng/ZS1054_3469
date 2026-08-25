package search

import (
	"path/filepath"
	"strings"
	"testing"

	"campusqa/internal/model"
	"campusqa/internal/store"
)

func TestSearchFiltersAndIndex(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "search.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	records := []model.Record{
		{ID: "r1", StudentID: "s1", Question: "exam schedule", Answer: "calendar", Category: "examination", Status: model.StatusPublished, Version: 1},
		{ID: "r2", StudentID: "s2", Question: "course enrollment", Answer: "portal", Category: "enrollment", Status: model.StatusDraft, Version: 1},
	}
	for _, record := range records {
		if err := database.PutRecord(record); err != nil {
			t.Fatal(err)
		}
	}
	results, err := NewService(database).Search(model.Query{Text: "exam", Status: model.StatusPublished})
	if err != nil || len(results) != 1 || results[0].Record.ID != "r1" {
		t.Fatalf("unexpected search: %v %#v", err, results)
	}
	index := NewIndex()
	index.Add(records[0])
	if got := index.Lookup("exam"); len(got) != 1 || got[0] != "r1" {
		t.Fatalf("unexpected index lookup %#v", got)
	}
	facets := BuildFacets(records)
	if len(facets) == 0 || len(Categories(results)) != 1 {
		t.Fatalf("unexpected facets %#v", facets)
	}
	if highlighted := Highlight(records[0], "exam"); !strings.Contains(highlighted["question"], "[exam]") {
		t.Fatal("highlight missing")
	}
	if FormatResult(results[0]) == "" || len(LimitResults(results, 1)) != 1 {
		t.Fatal("result formatting failed")
	}
}
