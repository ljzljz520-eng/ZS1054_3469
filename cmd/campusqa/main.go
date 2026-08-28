package main

import (
	"fmt"
	"os"
	"path/filepath"

	"campusqa/internal/flow018"
	"campusqa/internal/review"
	"campusqa/internal/store"
)

func main() {
	path := filepath.Join(".", "campusqa.db")
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	database, err := store.Open(path)
	if err != nil {
		fmt.Printf("campusqa: %v\n", err)
		return
	}
	defer database.Close()
	service := flow018.New(database)
	if len(os.Args) > 2 && os.Args[2] == "demo" {
		receipt, err := service.CreateReviewArchive(review.Submission{ID: "demo-001", StudentID: "student-001", Question: "How do I enroll in a course?", Answer: "Use the registration portal."}, "registrar")
		if err != nil {
			fmt.Printf("campusqa demo: %v\n", err)
			return
		}
		fmt.Printf("campusqa demo: %s %s %s\n", receipt.RecordID, receipt.Category, receipt.Status)
		return
	}
	fmt.Println("campusqa service ready; pass 'demo' to run a deterministic workflow")
}
