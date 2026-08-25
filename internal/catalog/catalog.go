package catalog

import (
	"sort"
	"strings"

	"campusqa/internal/model"
)

type Category struct {
	ID          string
	Name        string
	Description string
	Keywords    []string
	Priority    int
}

type Candidate struct {
	Category string
	Score    int
	Priority int
	Reason   string
}

type Catalog struct {
	categories []Category
}

func New() *Catalog {
	return &Catalog{categories: []Category{
		{ID: "enrollment", Name: "Enrollment", Description: "registration and course selection", Keywords: []string{"enroll", "registration", "course", "add", "drop"}, Priority: 20},
		{ID: "examination", Name: "Examination", Description: "tests and assessment schedules", Keywords: []string{"exam", "test", "assessment", "score", "grade"}, Priority: 30},
		{ID: "financial_aid", Name: "Financial Aid", Description: "scholarships and tuition support", Keywords: []string{"scholarship", "tuition", "grant", "aid"}, Priority: 10},
		{ID: "graduation", Name: "Graduation", Description: "graduation and completion", Keywords: []string{"graduate", "graduation", "degree", "completion"}, Priority: 40},
		{ID: "general", Name: "General", Description: "general academic guidance", Keywords: []string{}, Priority: 1},
	}}
}

func (c *Catalog) Categories() []Category {
	result := make([]Category, len(c.categories))
	copy(result, c.categories)
	return result
}

func (c *Catalog) Find(id string) (Category, bool) {
	for _, category := range c.categories {
		if category.ID == id {
			return category, true
		}
	}
	return Category{}, false
}

func (c *Catalog) Candidates(question string) []Candidate {
	normalized := model.NormalizeText(question)
	candidates := make([]Candidate, 0, len(c.categories))
	for _, category := range c.categories {
		score, reasons := scoreCategory(normalized, category)
		candidates = append(candidates, Candidate{Category: category.ID, Score: score, Priority: category.Priority, Reason: strings.Join(reasons, ",")})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].Priority > candidates[j].Priority
	})
	return candidates
}

func (c *Catalog) Classify(question string) string {
	candidates := c.Candidates(question)
	if len(candidates) == 0 || candidates[0].Score == 0 {
		return "general"
	}
	return candidates[0].Category
}

func scoreCategory(question string, category Category) (int, []string) {
	if len(category.Keywords) == 0 {
		return 0, nil
	}
	score := 0
	reasons := make([]string, 0)
	for _, keyword := range category.Keywords {
		if strings.Contains(question, keyword) {
			score++
			reasons = append(reasons, keyword)
		}
	}
	return score, reasons
}
