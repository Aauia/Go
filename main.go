package main

import (
	"fmt"
	"os"

	"github.com/Aauia/Go/internal/task"
)

func main() {
	svc := task.NewService(task.NewMemoryStore(), task.TextAnalyzer{})

	t, err := svc.Create("Hello Go world")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create:", err)
		os.Exit(1)
	}
	fmt.Printf("created %s (%s)\n", t.ID, t.Status)

	done, err := svc.Process(t.ID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "process:", err)
		os.Exit(1)
	}
	fmt.Printf("%s is %s: %+v\n", done.ID, done.Status, *done.Result)

	if _, err := svc.Create("   "); err != nil {
		fmt.Println("expected error:", err)
	}
}
