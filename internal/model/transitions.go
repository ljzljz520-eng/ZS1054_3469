package model

import "fmt"

func CanTransition(from, to RecordStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case StatusDraft:
		return to == StatusSubmitted
	case StatusSubmitted:
		return to == StatusApproved || to == StatusRejected
	case StatusApproved:
		return to == StatusPublished || to == StatusSubmitted
	case StatusPublished:
		return to == StatusArchived
	case StatusRejected:
		return to == StatusDraft || to == StatusSubmitted
	case StatusArchived:
		return false
	default:
		return false
	}
}

func Transition(r *Record, next RecordStatus) error {
	if r == nil {
		return fmt.Errorf("record is nil")
	}
	if !CanTransition(r.Status, next) {
		return fmt.Errorf("cannot transition %s to %s", r.Status, next)
	}
	r.Status = next
	r.Version++
	return nil
}

func StatusRank(status RecordStatus) int {
	switch status {
	case StatusDraft:
		return 1
	case StatusSubmitted:
		return 2
	case StatusApproved:
		return 3
	case StatusPublished:
		return 4
	case StatusArchived:
		return 5
	case StatusRejected:
		return 6
	default:
		return 0
	}
}

func IsTerminal(status RecordStatus) bool {
	return status == StatusArchived
}

func NextReviewState(approved bool) RecordStatus {
	if approved {
		return StatusApproved
	}
	return StatusRejected
}
