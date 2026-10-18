package task

import "fmt"

// Store is the storage abstraction used by the Service.
type Store interface {
	Save(t *Task)
	Get(id string) (*Task, error)
}

// MemoryStore keeps tasks in a map. Not safe for concurrent use
// (concurrency is a Checkpoint 2 topic).
type MemoryStore struct {
	tasks map[string]*Task
}

// NewMemoryStore returns an empty store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tasks: make(map[string]*Task)}
}

// Save stores (or replaces) the task under its ID.
func (s *MemoryStore) Save(t *Task) {
	s.tasks[t.ID] = t
}

// Get returns the stored task or ErrTaskNotFound.
func (s *MemoryStore) Get(id string) (*Task, error) {
	t, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrTaskNotFound, id)
	}
	return t, nil
}
