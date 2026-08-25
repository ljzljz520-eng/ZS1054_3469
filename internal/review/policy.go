package review

import (
	"sort"
	"strings"

	"campusqa/internal/model"
)

type Policy struct {
	Category      string
	MinimumAnswer int
	RequiredTags  []string
	EscalateWords []string
}

type PolicyResult struct {
	Allowed   bool
	Escalated bool
	Issues    []string
}

func DefaultPolicies() []Policy {
	return []Policy{
		{Category: "enrollment", MinimumAnswer: 8, RequiredTags: []string{"academic"}, EscalateWords: []string{"exception", "appeal"}},
		{Category: "examination", MinimumAnswer: 8, RequiredTags: []string{"academic"}, EscalateWords: []string{"retake", "appeal"}},
		{Category: "financial_aid", MinimumAnswer: 10, RequiredTags: []string{"support"}, EscalateWords: []string{"urgent", "appeal"}},
	}
}

func EvaluatePolicy(record model.Record, policies []Policy) PolicyResult {
	result := PolicyResult{Allowed: true, Issues: make([]string, 0)}
	policy, found := findPolicy(record.Category, policies)
	if !found {
		return result
	}
	if len(strings.TrimSpace(record.Answer)) < policy.MinimumAnswer {
		result.Allowed = false
		result.Issues = append(result.Issues, "answer is too short")
	}
	for _, required := range policy.RequiredTags {
		matched := false
		for _, tag := range record.Tags {
			if strings.EqualFold(tag, required) {
				matched = true
				break
			}
		}
		if !matched {
			result.Allowed = false
			result.Issues = append(result.Issues, "missing tag: "+required)
		}
	}
	for _, word := range policy.EscalateWords {
		if strings.Contains(strings.ToLower(record.Question), word) || strings.Contains(strings.ToLower(record.Answer), word) {
			result.Escalated = true
		}
	}
	return result
}

func findPolicy(category string, policies []Policy) (Policy, bool) {
	for _, policy := range policies {
		if policy.Category == category {
			return policy, true
		}
	}
	return Policy{}, false
}

func QueuePriority(record model.Record, result PolicyResult) int {
	priority := model.StatusRank(record.Status) * 10
	if result.Escalated {
		priority += 50
	}
	if !result.Allowed {
		priority += len(result.Issues) * 5
	}
	return priority
}

func RankQueue(records []model.Record, policies []Policy) []model.Record {
	result := make([]model.Record, len(records))
	copy(result, records)
	sort.SliceStable(result, func(i, j int) bool {
		left := EvaluatePolicy(result[i], policies)
		right := EvaluatePolicy(result[j], policies)
		lp := QueuePriority(result[i], left)
		rp := QueuePriority(result[j], right)
		if lp != rp {
			return lp > rp
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func PolicyNames(policies []Policy) []string {
	names := make([]string, 0, len(policies))
	for _, policy := range policies {
		names = append(names, policy.Category)
	}
	sort.Strings(names)
	return names
}
