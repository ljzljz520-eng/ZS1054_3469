package report

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"campusqa/internal/model"
)

type Importer struct {
	seen map[string]bool
}

type ImportSummary struct {
	Accepted int
	Rejected int
	Issues   []string
}

func NewImporter() *Importer {
	return &Importer{seen: make(map[string]bool)}
}

func (i *Importer) CSV(input io.Reader) ([]model.Record, ImportSummary, error) {
	if input == nil {
		return nil, ImportSummary{}, errors.New("csv input is required")
	}
	reader := csv.NewReader(bufio.NewReader(input))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, ImportSummary{}, fmt.Errorf("read csv: %w", err)
	}
	if len(rows) == 0 {
		return nil, ImportSummary{}, errors.New("csv has no rows")
	}
	start := 0
	if strings.EqualFold(strings.TrimSpace(rows[0][0]), "id") {
		start = 1
	}
	records := make([]model.Record, 0)
	summary := ImportSummary{Issues: make([]string, 0)}
	for index := start; index < len(rows); index++ {
		record, err := parseRow(rows[index])
		if err != nil {
			summary.Rejected++
			summary.Issues = append(summary.Issues, fmt.Sprintf("row %d: %v", index+1, err))
			continue
		}
		if i.seen[record.ID] {
			summary.Rejected++
			summary.Issues = append(summary.Issues, fmt.Sprintf("row %d: duplicate id", index+1))
			continue
		}
		i.seen[record.ID] = true
		record.Status = model.StatusDraft
		record.Version = 1
		record.Tags = model.NormalizeTags(record.Tags)
		record.Category = "general"
		records = append(records, record)
		summary.Accepted++
	}
	return records, summary, nil
}

func parseRow(row []string) (model.Record, error) {
	if len(row) < 5 {
		return model.Record{}, errors.New("expected id, student, question, answer, category")
	}
	record := model.Record{ID: strings.TrimSpace(row[0]), StudentID: strings.TrimSpace(row[1]), Question: strings.TrimSpace(row[2]), Answer: strings.TrimSpace(row[3]), Category: strings.TrimSpace(row[4]), Status: model.StatusDraft, Version: 1}
	if record.Category == "" {
		record.Category = "general"
	}
	if err := record.Validate(); err != nil {
		return model.Record{}, err
	}
	return record, nil
}

func ParseLines(input string) ([]model.Record, ImportSummary, error) {
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	rows := make([][]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		rows = append(rows, strings.Split(line, "|"))
	}
	importer := NewImporter()
	records := make([]model.Record, 0)
	summary := ImportSummary{Issues: make([]string, 0)}
	for index, row := range rows {
		record, err := parseRow(row)
		if err != nil {
			summary.Rejected++
			summary.Issues = append(summary.Issues, fmt.Sprintf("line %d: %v", index+1, err))
			continue
		}
		if importer.seen[record.ID] {
			summary.Rejected++
			continue
		}
		importer.seen[record.ID] = true
		records = append(records, record)
		summary.Accepted++
	}
	return records, summary, nil
}
