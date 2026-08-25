package catalog

import "testing"

func TestClassificationAndLabels(t *testing.T) {
	catalog := New()
	if got := catalog.Classify("How do I enroll in a course?"); got != "enrollment" {
		t.Fatalf("unexpected category %s", got)
	}
	candidates := catalog.Candidates("exam score")
	if len(candidates) == 0 || candidates[0].Category != "examination" {
		t.Fatalf("unexpected candidates %#v", candidates)
	}
	labels := TopLabels([]string{"course exam", "course guide", "exam policy"}, 2)
	if len(labels) != 2 || labels[0].Value != "course" {
		t.Fatalf("unexpected labels %#v", labels)
	}
	explanation := catalog.Explain("course exam")
	if explanation.Category == "" || len(catalog.Suggestions("course", 2)) != 2 {
		t.Fatalf("unexpected explanation %#v", explanation)
	}
	if ResolveAlias("registration") != "enrollment" {
		t.Fatal("alias was not resolved")
	}
}

func TestRulesAndCategoryLookup(t *testing.T) {
	catalog := New()
	if err := ValidateCategory(catalog, "graduation"); err != nil {
		t.Fatal(err)
	}
	ok, issues := DefaultRules().Evaluate("course enrollment", "enrollment")
	if !ok || len(issues) != 0 {
		t.Fatalf("unexpected rule result %v %#v", ok, issues)
	}
}
