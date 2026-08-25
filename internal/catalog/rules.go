package catalog

import (
	"fmt"
	"strings"

	"campusqa/internal/model"
)

type Rule struct {
	Name      string
	Category  string
	Required  []string
	Forbidden []string
}

type RuleSet struct {
	rules []Rule
}

func DefaultRules() RuleSet {
	return RuleSet{rules: []Rule{
		{Name: "enrollment-detail", Category: "enrollment", Required: []string{"course"}},
		{Name: "exam-detail", Category: "examination", Required: []string{"exam"}},
		{Name: "aid-detail", Category: "financial_aid", Required: []string{"aid"}},
	}}
}

func (r RuleSet) Evaluate(question, category string) (bool, []string) {
	normalized := model.NormalizeText(question)
	issues := make([]string, 0)
	for _, rule := range r.rules {
		if rule.Category != category {
			continue
		}
		for _, required := range rule.Required {
			if !strings.Contains(normalized, required) {
				issues = append(issues, fmt.Sprintf("missing %s", required))
			}
		}
		for _, forbidden := range rule.Forbidden {
			if strings.Contains(normalized, forbidden) {
				issues = append(issues, fmt.Sprintf("contains %s", forbidden))
			}
		}
	}
	return len(issues) == 0, issues
}

func (r RuleSet) Names() []string {
	names := make([]string, 0, len(r.rules))
	for _, rule := range r.rules {
		names = append(names, rule.Name)
	}
	return names
}

func ValidateCategory(catalog *Catalog, category string) error {
	if catalog == nil {
		return fmt.Errorf("catalog is nil")
	}
	if _, ok := catalog.Find(category); !ok {
		return fmt.Errorf("unknown category %s", category)
	}
	return nil
}
