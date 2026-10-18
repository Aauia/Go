package task

import "fmt"

// Service coordinates creating, retrieving and processing tasks.
type Service struct {
	store     Store
	processor Processor
	nextID    int
}

// NewService builds a Service from its dependencies.
func NewService(store Store, processor Processor) *Service {
	return &Service{store: store, processor: processor}
}

// Create validates text, assigns a unique ID and stores a pending task.
func (s *Service) Create(text string) (Task, error) {
	if err := ValidateText(text); err != nil {
		return Task{}, err
	}
	s.nextID++
	t := &Task{
		ID:     fmt.Sprintf("task-%d", s.nextID),
		Text:   text,
		Status: StatusPending,
	}
	s.store.Save(t)
	return *t, nil
}

// Get returns a copy of the task with the given ID, so callers cannot
// modify the stored task by accident.
func (s *Service) Get(id string) (Task, error) {
	t, err := s.store.Get(id)
	if err != nil {
		return Task{}, err
	}
	return copyTask(t), nil
}

// Process runs the task through pending -> processing -> completed (or failed).
func (s *Service) Process(id string) (Task, error) {
	t, err := s.store.Get(id)
	if err != nil {
		return Task{}, err
	}

	if err := t.transition(StatusProcessing); err != nil {
		return Task{}, err
	}

	res, err := s.processor.Process(t.Text)
	if err != nil {
		// transition from processing to failed is always allowed.
		_ = t.transition(StatusFailed)
		t.Err = err.Error()
		return copyTask(t), fmt.Errorf("process %s: %w", t.ID, err)
	}

	t.Result = &res
	if err := t.transition(StatusCompleted); err != nil {
		return Task{}, err
	}
	return copyTask(t), nil
}

// copyTask copies a task including the Result it points to.
func copyTask(t *Task) Task {
	c := *t
	if t.Result != nil {
		r := *t.Result
		c.Result = &r
	}
	return c
}
