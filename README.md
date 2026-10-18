# SIS – Text Analyzer (Checkpoint 1)

A synchronous, in-memory Go application. A **task** is a unit of work: analyse
a piece of text and produce statistics about it.

```
Input:  "Hello Go world"
Result: words=3, characters=14, lines=1, uniqueWords=3, longestWord="Hello"
```

## Run

```bash
go run .
gofmt -l .
go vet ./...
go test ./...
```

## Structure

| File | Purpose |
|------|---------|
| `internal/task/task.go` | `Task`, `Result`, `Status`, lifecycle rules, errors, validation |
| `internal/task/processor.go` | `Processor` interface and `TextAnalyzer` implementation |
| `internal/task/store.go` | `Store` interface and map-based `MemoryStore` |
| `internal/task/service.go` | `Service`: create / get / process tasks |
| `internal/task/task_test.go` | Unit tests |
| `main.go` | Small demo |

## Validation

Text that is empty or contains only whitespace is rejected with `ErrEmptyText`.

## Task lifecycle

```
pending ──► processing ──► completed
                 └───────► failed
```

| From | Allowed next states |
|------|---------------------|
| pending | processing |
| processing | completed, failed |
| completed | – (terminal) |
| failed | – (terminal) |

Any other transition (for example processing an already completed task)
returns `ErrInvalidTransition` and leaves the task unchanged.

## Errors

Expected errors are returned, never panicked: `ErrEmptyText`, `ErrTaskNotFound`,
`ErrInvalidTransition`. Check them with `errors.Is`.
