package review

import (
	"sort"

	"campusqa/internal/model"
)

type Queue struct {
	items []model.Record
}

func NewQueue(records []model.Record) *Queue {
	items := make([]model.Record, len(records))
	copy(items, records)
	return &Queue{items: items}
}

func (q *Queue) Pending() []model.Record {
	result := make([]model.Record, 0)
	for _, record := range q.items {
		if record.Status == model.StatusSubmitted {
			result = append(result, record.Clone())
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Version != result[j].Version {
			return result[i].Version < result[j].Version
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func (q *Queue) Add(record model.Record) {
	for index := range q.items {
		if q.items[index].ID == record.ID {
			q.items[index] = record
			return
		}
	}
	q.items = append(q.items, record)
}

func (q *Queue) Remove(id string) bool {
	for index, record := range q.items {
		if record.ID == id {
			q.items = append(q.items[:index], q.items[index+1:]...)
			return true
		}
	}
	return false
}

func (q *Queue) Size() int {
	return len(q.Pending())
}

func (q *Queue) Snapshot() []model.Record {
	result := make([]model.Record, len(q.items))
	for i, record := range q.items {
		result[i] = record.Clone()
	}
	return result
}
