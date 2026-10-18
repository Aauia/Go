package task

import (
	"errors"
	"testing"
)

func newTestService() *Service {
	return NewService(NewMemoryStore(), TextAnalyzer{})
}

// failingProcessor always fails; it demonstrates the Processor interface.
type failingProcessor struct{}

func (failingProcessor) Process(string) (Result, error) {
	return Result{}, errors.New("boom")
}

func TestCreate_Success(t *testing.T) {
	svc := newTestService()
	got, err := svc.Create("Hello Go world")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if got.ID == "" {
		t.Error("expected non-empty ID")
	}
	if got.Status != StatusPending {
		t.Errorf("status = %q, want %q", got.Status, StatusPending)
	}
	if got.Text != "Hello Go world" {
		t.Errorf("text = %q", got.Text)
	}
	if got.Result != nil {
		t.Error("new task must not have a result")
	}
}

func TestCreate_InvalidInput(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"spaces", "     "},
		{"whitespace mix", " \t\n "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()
			_, err := svc.Create(tt.text)
			if !errors.Is(err, ErrEmptyText) {
				t.Errorf("err = %v, want ErrEmptyText", err)
			}
		})
	}
}

func TestGet_Success(t *testing.T) {
	svc := newTestService()
	created, _ := svc.Create("some text")

	got, err := svc.Get(created.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if got.ID != created.ID || got.Text != created.Text {
		t.Errorf("got %+v, want %+v", got, created)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.Get("does-not-exist")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestProcess_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.Process("nope")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Errorf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestProcess_SuccessAndState(t *testing.T) {
	svc := newTestService()
	created, _ := svc.Create("Hello Go world")

	done, err := svc.Process(created.ID)
	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}
	if done.Status != StatusCompleted {
		t.Errorf("status = %q, want %q", done.Status, StatusCompleted)
	}
	if done.Result == nil {
		t.Fatal("result is nil")
	}

	// The stored task must reflect the processing too.
	stored, _ := svc.Get(created.ID)
	if stored.Status != StatusCompleted || stored.Result == nil {
		t.Errorf("stored task not updated: %+v", stored)
	}
}

func TestTextAnalyzer_Results(t *testing.T) {
	tests := []struct {
		name string
		text string
		want Result
	}{
		{
			name: "simple",
			text: "Hello Go world",
			want: Result{Words: 3, Characters: 14, Lines: 1, UniqueWords: 3, LongestWord: "Hello"},
		},
		{
			name: "case-insensitive unique words and punctuation",
			text: "Go, go! GO?",
			want: Result{Words: 3, Characters: 11, Lines: 1, UniqueWords: 1, LongestWord: "Go"},
		},
		{
			name: "multiple lines with blank line",
			text: "one two\n\nthree",
			want: Result{Words: 3, Characters: 14, Lines: 2, UniqueWords: 3, LongestWord: "three"},
		},
		{
			name: "unicode counts runes not bytes",
			text: "Привет мир",
			want: Result{Words: 2, Characters: 10, Lines: 1, UniqueWords: 2, LongestWord: "Привет"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()
			created, err := svc.Create(tt.text)
			if err != nil {
				t.Fatal(err)
			}
			done, err := svc.Process(created.ID)
			if err != nil {
				t.Fatal(err)
			}
			if *done.Result != tt.want {
				t.Errorf("result = %+v, want %+v", *done.Result, tt.want)
			}
		})
	}
}

func TestProcess_InvalidTransition(t *testing.T) {
	svc := newTestService()
	created, _ := svc.Create("hello")
	if _, err := svc.Process(created.ID); err != nil {
		t.Fatal(err)
	}

	_, err := svc.Process(created.ID) // already completed
	if !errors.Is(err, ErrInvalidTransition) {
		t.Errorf("err = %v, want ErrInvalidTransition", err)
	}
	stored, _ := svc.Get(created.ID)
	if stored.Status != StatusCompleted {
		t.Errorf("status changed to %q after invalid operation", stored.Status)
	}
}

func TestTransition_Table(t *testing.T) {
	tests := []struct {
		from, to Status
		wantErr  bool
	}{
		{StatusPending, StatusProcessing, false},
		{StatusProcessing, StatusCompleted, false},
		{StatusProcessing, StatusFailed, false},
		{StatusPending, StatusCompleted, true},
		{StatusCompleted, StatusProcessing, true},
		{StatusFailed, StatusProcessing, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.from)+"->"+string(tt.to), func(t *testing.T) {
			tk := &Task{ID: "x", Status: tt.from}
			err := tk.transition(tt.to)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrInvalidTransition) {
				t.Errorf("err = %v, want ErrInvalidTransition", err)
			}
			wantStatus := tt.to
			if tt.wantErr {
				wantStatus = tt.from
			}
			if tk.Status != wantStatus {
				t.Errorf("status = %q, want %q", tk.Status, wantStatus)
			}
		})
	}
}

func TestProcess_ProcessorFailure(t *testing.T) {
	svc := NewService(NewMemoryStore(), failingProcessor{})
	created, _ := svc.Create("hello")

	_, err := svc.Process(created.ID)
	if err == nil {
		t.Fatal("expected error")
	}
	stored, _ := svc.Get(created.ID)
	if stored.Status != StatusFailed {
		t.Errorf("status = %q, want %q", stored.Status, StatusFailed)
	}
	if stored.Result != nil || stored.Err == "" {
		t.Errorf("unexpected stored task: %+v", stored)
	}
}

func TestMultipleIndependentTasks(t *testing.T) {
	svc := newTestService()
	a, _ := svc.Create("alpha beta")
	b, _ := svc.Create("one two three four")
	c, _ := svc.Create("untouched")

	if a.ID == b.ID || b.ID == c.ID || a.ID == c.ID {
		t.Fatalf("IDs not unique: %s %s %s", a.ID, b.ID, c.ID)
	}

	if _, err := svc.Process(a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Process(b.ID); err != nil {
		t.Fatal(err)
	}

	gotA, _ := svc.Get(a.ID)
	gotB, _ := svc.Get(b.ID)
	gotC, _ := svc.Get(c.ID)

	if gotA.Result.Words != 2 || gotA.Text != "alpha beta" {
		t.Errorf("task A mixed up: %+v", gotA)
	}
	if gotB.Result.Words != 4 || gotB.Text != "one two three four" {
		t.Errorf("task B mixed up: %+v", gotB)
	}
	if gotC.Status != StatusPending || gotC.Result != nil {
		t.Errorf("task C should be untouched: %+v", gotC)
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	svc := newTestService()
	created, _ := svc.Create("hello world")
	svc.Process(created.ID)

	got, _ := svc.Get(created.ID)
	got.Status = StatusPending
	got.Result.Words = 999

	again, _ := svc.Get(created.ID)
	if again.Status != StatusCompleted || again.Result.Words != 2 {
		t.Errorf("modifying a returned copy changed stored task: %+v", again)
	}
}
