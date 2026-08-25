package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"

	"campusqa/internal/model"
)

type Filter struct {
	Categories []string
	Statuses   []model.RecordStatus
	Student    string
	Text       string
}

func ApplyFilter(records []model.Record, filter Filter) []model.Record {
	result := make([]model.Record, 0)
	for _, record := range records {
		if len(filter.Categories) > 0 && !containsString(filter.Categories, record.Category) {
			continue
		}
		if len(filter.Statuses) > 0 && !containsStatus(filter.Statuses, record.Status) {
			continue
		}
		if filter.Student != "" && record.StudentID != filter.Student {
			continue
		}
		if filter.Text != "" && !strings.Contains(model.NormalizeText(record.Question+" "+record.Answer), model.NormalizeText(filter.Text)) {
			continue
		}
		result = append(result, record.Clone())
	}
	return result
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func containsStatus(values []model.RecordStatus, value model.RecordStatus) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func SortForExport(records []model.Record) []model.Record {
	result := model.CopyRecords(records)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].StudentID != result[j].StudentID {
			return result[i].StudentID < result[j].StudentID
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func WriteCSV(writer io.Writer, records []model.Record) error {
	if writer == nil {
		return fmt.Errorf("writer is required")
	}
	output := csv.NewWriter(writer)
	if err := output.Write([]string{"id", "student_id", "question", "answer", "category", "status", "version"}); err != nil {
		return err
	}
	for _, record := range SortForExport(records) {
		if err := output.Write([]string{record.ID, record.StudentID, record.Question, record.Answer, record.Category, string(record.Status), fmt.Sprintf("%d", record.Version)}); err != nil {
			return err
		}
	}
	output.Flush()
	return output.Error()
}

func Markdown(summary Summary) string {
	lines := []string{"# Campus Academic Q&A", "", fmt.Sprintf("Total records: %d", summary.Total), "", "| Category | Count |", "| --- | ---: |"}
	keys := make([]string, 0, len(summary.ByCategory))
	for key := range summary.ByCategory {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("| %s | %d |", key, summary.ByCategory[key]))
	}
	return strings.Join(lines, "\n")
}
