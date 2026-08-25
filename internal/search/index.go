package search

import (
	"sort"
	"strings"

	"campusqa/internal/model"
)

type Index struct {
	terms map[string]map[string]bool
}

func NewIndex() *Index {
	return &Index{terms: make(map[string]map[string]bool)}
}

func (i *Index) Add(record model.Record) {
	if i.terms == nil {
		i.terms = make(map[string]map[string]bool)
	}
	for _, term := range terms(record) {
		if i.terms[term] == nil {
			i.terms[term] = make(map[string]bool)
		}
		i.terms[term][record.ID] = true
	}
}

func (i *Index) Remove(record model.Record) {
	for _, term := range terms(record) {
		if ids := i.terms[term]; ids != nil {
			delete(ids, record.ID)
			if len(ids) == 0 {
				delete(i.terms, term)
			}
		}
	}
}

func (i *Index) Lookup(text string) []string {
	termList := strings.Fields(model.NormalizeText(text))
	if len(termList) == 0 {
		return nil
	}
	counts := make(map[string]int)
	for _, term := range termList {
		for id := range i.terms[term] {
			counts[id]++
		}
	}
	ids := make([]string, 0, len(counts))
	for id := range counts {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(a, b int) bool {
		if counts[ids[a]] != counts[ids[b]] {
			return counts[ids[a]] > counts[ids[b]]
		}
		return ids[a] < ids[b]
	})
	return ids
}

func (i *Index) Size() int {
	return len(i.terms)
}

func terms(record model.Record) []string {
	all := strings.Join([]string{record.Question, record.Answer, record.Category, strings.Join(record.Tags, " ")}, " ")
	words := strings.Fields(model.NormalizeText(all))
	seen := make(map[string]bool)
	result := make([]string, 0, len(words))
	for _, word := range words {
		if seen[word] {
			continue
		}
		seen[word] = true
		result = append(result, word)
	}
	return result
}
