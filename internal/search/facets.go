package search

import (
	"fmt"
	"sort"
	"strings"

	"campusqa/internal/model"
)

type Facet struct {
	Name  string
	Value string
	Count int
}

func BuildFacets(records []model.Record) []Facet {
	counts := make(map[string]int)
	for _, record := range records {
		counts["category:"+record.Category]++
		counts["status:"+string(record.Status)]++
		counts["student:"+record.StudentID]++
	}
	facets := make([]Facet, 0, len(counts))
	for key, count := range counts {
		parts := strings.SplitN(key, ":", 2)
		facets = append(facets, Facet{Name: parts[0], Value: parts[1], Count: count})
	}
	sort.Slice(facets, func(i, j int) bool {
		if facets[i].Count != facets[j].Count {
			return facets[i].Count > facets[j].Count
		}
		if facets[i].Name != facets[j].Name {
			return facets[i].Name < facets[j].Name
		}
		return facets[i].Value < facets[j].Value
	})
	return facets
}

func Highlight(record model.Record, text string) map[string]string {
	needle := strings.TrimSpace(text)
	fields := map[string]string{"question": record.Question, "answer": record.Answer, "category": record.Category}
	if needle == "" {
		return fields
	}
	for key, value := range fields {
		fields[key] = strings.ReplaceAll(value, needle, "["+needle+"]")
	}
	return fields
}

func FormatResult(result Result) string {
	return fmt.Sprintf("%s|%s|%d|%s", result.Record.ID, result.Record.Category, result.Score, strings.Join(result.Match, ","))
}

func LimitResults(results []Result, limit int) []Result {
	if limit <= 0 || limit >= len(results) {
		return results
	}
	return results[:limit]
}

func Categories(results []Result) []string {
	set := make(map[string]bool)
	for _, result := range results {
		set[result.Record.Category] = true
	}
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}
