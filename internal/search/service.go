package search

import (
	"sort"
	"strings"

	"campusqa/internal/model"
	"campusqa/internal/store"
)

type Service struct {
	store *store.Store
}

type Result struct {
	Record model.Record
	Score  int
	Match  []string
}

func NewService(database *store.Store) *Service {
	return &Service{store: database}
}

func (s *Service) Search(query model.Query) ([]Result, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	records, err := s.store.ListRecords()
	if err != nil {
		return nil, err
	}
	needle := model.NormalizeText(query.Text)
	results := make([]Result, 0)
	for _, record := range records {
		if query.StudentID != "" && record.StudentID != query.StudentID {
			continue
		}
		if query.Category != "" && record.Category != query.Category {
			continue
		}
		if query.Status != "" && record.Status != query.Status {
			continue
		}
		score, matches := scoreRecord(record, needle)
		if needle != "" && score == 0 {
			continue
		}
		results = append(results, Result{Record: record, Score: score, Match: matches})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Record.ID < results[j].Record.ID
	})
	if query.Limit > 0 && query.Limit < len(results) {
		results = results[:query.Limit]
	}
	return results, nil
}

func scoreRecord(record model.Record, needle string) (int, []string) {
	if needle == "" {
		return 0, nil
	}
	score := 0
	matches := make([]string, 0)
	fields := []struct {
		name   string
		value  string
		weight int
	}{
		{"question", record.Question, 5},
		{"answer", record.Answer, 3},
		{"category", record.Category, 2},
		{"tags", strings.Join(record.Tags, " "), 1},
	}
	for _, field := range fields {
		if strings.Contains(model.NormalizeText(field.value), needle) {
			score += field.weight
			matches = append(matches, field.name)
		}
	}
	return score, matches
}

func FilterByStatus(records []model.Record, status model.RecordStatus) []model.Record {
	filtered := make([]model.Record, 0)
	for _, record := range records {
		if status == "" || record.Status == status {
			filtered = append(filtered, record.Clone())
		}
	}
	return filtered
}

func Summarize(results []Result) map[string]int {
	summary := make(map[string]int)
	for _, result := range results {
		summary[result.Record.Category]++
	}
	return summary
}
