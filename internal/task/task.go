// Package task contains the domain model, processing logic and in-memory
// storage for the SIS text-analysis tasks.
package task

import (
	"errors"
	"fmt"
	"strings"
)

// Status is the lifecycle state of a Task.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// Sentinel errors. Callers can check them with errors.Is.
var (
	ErrEmptyText         = errors.New("task: text is empty or whitespace-only")
	ErrTaskNotFound      = errors.New("task: not found")
	ErrInvalidTransition = errors.New("task: invalid state transition")
)

// allowedTransitions describes the lifecycle:
//
//	pending -> processing -> completed
//	                      \-> failed
//
// completed and failed are terminal states.
var allowedTransitions = map[Status][]Status{
	StatusPending:    {StatusProcessing},
	StatusProcessing: {StatusCompleted, StatusFailed},
}

// Result is the output of analysing a text. It contains only value types,
// so copying a Task never shares mutable data between copies.
type Result struct {
	Words       int    // number of whitespace-separated words
	Characters  int    // number of runes (not bytes)
	Lines       int    // number of non-empty lines
	UniqueWords int    // distinct words, case-insensitive
	LongestWord string // first longest word, punctuation trimmed
}

// Task is a unit of work: analyse a piece of text.
type Task struct {
	ID     string
	Text   string  // input
	Status Status  // current lifecycle state
	Result *Result // nil until the task has completed
	Err    string  // reason of failure, empty otherwise
}

// ValidateText checks that the input of a task is acceptable.
func ValidateText(text string) error {
	if strings.TrimSpace(text) == "" {
		return ErrEmptyText
	}
	return nil
}

// transition moves the task to the next status if the lifecycle allows it.
// It has a pointer receiver because it modifies the task.
func (t *Task) transition(next Status) error {
	for _, allowed := range allowedTransitions[t.Status] {
		if allowed == next {
			t.Status = next
			return nil
		}
	}
	return fmt.Errorf("%w: %s -> %s (task %s)", ErrInvalidTransition, t.Status, next, t.ID)
}
