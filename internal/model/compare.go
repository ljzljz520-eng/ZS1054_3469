package model

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

type Change struct {
	Field string
	From  string
	To    string
}

func CompareRecords(before, after Record) []Change {
	changes := make([]Change, 0)
	if before.StudentID != after.StudentID {
		changes = append(changes, Change{Field: "student_id", From: before.StudentID, To: after.StudentID})
	}
	if before.Question != after.Question {
		changes = append(changes, Change{Field: "question", From: before.Question, To: after.Question})
	}
	if before.Category != after.Category {
		changes = append(changes, Change{Field: "category", From: before.Category, To: after.Category})
	}
	if before.Answer != after.Answer {
		changes = append(changes, Change{Field: "answer", From: before.Answer, To: after.Answer})
	}
	if before.Status != after.Status {
		changes = append(changes, Change{Field: "status", From: string(before.Status), To: string(after.Status)})
	}
	if !sameStrings(before.Tags, after.Tags) {
		changes = append(changes, Change{Field: "tags", From: strings.Join(before.Tags, ","), To: strings.Join(after.Tags, ",")})
	}
	return changes
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func RecordFingerprint(record Record) string {
	tags := append([]string(nil), record.Tags...)
	sort.Strings(tags)
	value := strings.Join([]string{record.ID, record.StudentID, record.Question, record.Category, record.Answer, string(record.Status), strings.Join(tags, ",")}, "|")
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func SortRecords(records []Record, newestFirst bool) []Record {
	result := make([]Record, len(records))
	copy(result, records)
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Version != result[j].Version {
			if newestFirst {
				return result[i].Version > result[j].Version
			}
			return result[i].Version < result[j].Version
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func CopyRecords(records []Record) []Record {
	result := make([]Record, len(records))
	for index, record := range records {
		result[index] = record.Clone()
	}
	return result
}
