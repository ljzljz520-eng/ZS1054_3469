package catalog

import (
	"sort"
	"strings"

	"campusqa/internal/model"
)

type Explanation struct {
	Category   string
	Score      int
	Reasons    []string
	Alternates []string
}

type Suggestion struct {
	Category string
	Prompt   string
	Weight   int
}

func (c *Catalog) Explain(question string) Explanation {
	candidates := c.Candidates(question)
	if len(candidates) == 0 {
		return Explanation{Category: "general", Reasons: []string{"no categories configured"}}
	}
	selected := candidates[0]
	if selected.Score == 0 {
		return Explanation{Category: "general", Score: 0, Reasons: []string{"no keyword matched"}}
	}
	alternates := make([]string, 0)
	for _, candidate := range candidates[1:] {
		if candidate.Score == selected.Score {
			alternates = append(alternates, candidate.Category)
		}
	}
	sort.Strings(alternates)
	reasons := strings.Split(selected.Reason, ",")
	if selected.Reason == "" {
		reasons = nil
	}
	return Explanation{Category: selected.Category, Score: selected.Score, Reasons: reasons, Alternates: alternates}
}

func (c *Catalog) Suggestions(question string, limit int) []Suggestion {
	if limit <= 0 {
		return nil
	}
	text := model.NormalizeText(question)
	suggestions := make([]Suggestion, 0)
	for _, category := range c.categories {
		if category.ID == "general" {
			continue
		}
		weight := 0
		for _, keyword := range category.Keywords {
			if strings.Contains(text, keyword) {
				weight += category.Priority
			}
		}
		if weight == 0 {
			weight = category.Priority / 10
		}
		suggestions = append(suggestions, Suggestion{Category: category.ID, Prompt: "Ask about " + category.Name, Weight: weight})
	}
	sort.SliceStable(suggestions, func(i, j int) bool {
		if suggestions[i].Weight != suggestions[j].Weight {
			return suggestions[i].Weight > suggestions[j].Weight
		}
		return suggestions[i].Category < suggestions[j].Category
	})
	if limit > len(suggestions) {
		limit = len(suggestions)
	}
	return suggestions[:limit]
}

func ResolveAlias(value string) string {
	aliases := map[string]string{"registration": "enrollment", "tests": "examination", "scholarships": "financial_aid", "degree": "graduation"}
	key := strings.ToLower(strings.TrimSpace(value))
	if resolved, ok := aliases[key]; ok {
		return resolved
	}
	return key
}
