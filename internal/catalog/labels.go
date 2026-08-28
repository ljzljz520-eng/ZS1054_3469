package catalog

import (
	"sort"
	"strings"
)

type Label struct {
	Value string
	Count int
}

func BuildLabels(texts []string) []Label {
	counts := make(map[string]int)
	for _, text := range texts {
		for _, word := range strings.Fields(strings.ToLower(text)) {
			if len(word) < 3 {
				continue
			}
			counts[word]++
		}
	}
	labels := make([]Label, 0, len(counts))
	for value, count := range counts {
		labels = append(labels, Label{Value: value, Count: count})
	}
	sort.Slice(labels, func(i, j int) bool {
		if labels[i].Count != labels[j].Count {
			return labels[i].Count > labels[j].Count
		}
		return labels[i].Value < labels[j].Value
	})
	return labels
}

func TopLabels(texts []string, limit int) []Label {
	labels := BuildLabels(texts)
	if limit < 0 {
		return nil
	}
	if limit > len(labels) {
		limit = len(labels)
	}
	return labels[:limit]
}

func MergeLabels(primary, secondary []string) []string {
	values := append(append([]string{}, primary...), secondary...)
	seen := make(map[string]bool)
	result := make([]string, 0, len(values))
	for _, value := range values {
		key := strings.TrimSpace(strings.ToLower(value))
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, key)
	}
	return result
}
