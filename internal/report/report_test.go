package report

import (
	"strings"
	"testing"
)

func TestImportAndReport(t *testing.T) {
	records, imported, err := ParseLines("r1|s1|course enrollment?|portal|general\nr2|s2|exam date?|calendar|examination")
	if err != nil || imported.Accepted != 2 || len(records) != 2 {
		t.Fatalf("import failed: %v %#v", err, imported)
	}
	summary := Build(records)
	if summary.Total != 2 || summary.ByCategory["general"] != 1 {
		t.Fatalf("unexpected summary %#v", summary)
	}
	if Render(summary) == "" {
		t.Fatal("empty report")
	}
	filtered := ApplyFilter(records, Filter{Categories: []string{"general"}})
	if len(filtered) != 1 || len(SortForExport(records)) != 2 {
		t.Fatalf("unexpected filtered records %#v", filtered)
	}
	var csv strings.Builder
	if err := WriteCSV(&csv, records); err != nil || !strings.Contains(csv.String(), "student_id") {
		t.Fatal(err)
	}
	if !strings.Contains(Markdown(summary), "Campus Academic") {
		t.Fatal("markdown report missing title")
	}
	if len(GroupByStudent(records)["s1"]) != 1 {
		t.Fatal("student grouping failed")
	}
}
