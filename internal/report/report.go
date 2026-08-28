package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"campusqa/internal/model"
)

type Summary struct {
	Total      int            `json:"total"`
	ByCategory map[string]int `json:"by_category"`
	ByStatus   map[string]int `json:"by_status"`
	Students   []string       `json:"students"`
}

func Build(records []model.Record) Summary {
	summary := Summary{ByCategory: make(map[string]int), ByStatus: make(map[string]int), Students: make([]string, 0)}
	studentSet := make(map[string]bool)
	for _, record := range records {
		summary.Total++
		summary.ByCategory[record.Category]++
		summary.ByStatus[string(record.Status)]++
		if !studentSet[record.StudentID] {
			studentSet[record.StudentID] = true
			summary.Students = append(summary.Students, record.StudentID)
		}
	}
	sort.Strings(summary.Students)
	return summary
}

func Render(summary Summary) string {
	parts := make([]string, 0, len(summary.ByCategory))
	for key, value := range summary.ByCategory {
		parts = append(parts, fmt.Sprintf("%s=%d", key, value))
	}
	sort.Strings(parts)
	statuses := make([]string, 0, len(summary.ByStatus))
	for key, value := range summary.ByStatus {
		statuses = append(statuses, fmt.Sprintf("%s=%d", key, value))
	}
	sort.Strings(statuses)
	return fmt.Sprintf("total=%d categories[%s] statuses[%s] students[%s]", summary.Total, strings.Join(parts, ","), strings.Join(statuses, ","), strings.Join(summary.Students, ","))
}

func JSON(summary Summary) ([]byte, error) {
	return json.MarshalIndent(summary, "", "  ")
}

func GroupByStudent(records []model.Record) map[string][]model.Record {
	groups := make(map[string][]model.Record)
	for _, record := range records {
		groups[record.StudentID] = append(groups[record.StudentID], record.Clone())
	}
	for student := range groups {
		sort.Slice(groups[student], func(i, j int) bool { return groups[student][i].ID < groups[student][j].ID })
	}
	return groups
}
