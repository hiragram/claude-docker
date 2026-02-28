package pipeline

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// mockStage is a test helper that records its execution.
type mockStage struct {
	name string
	err  error
	ran  bool
}

func (s *mockStage) Name() string { return s.name }
func (s *mockStage) Run(_ context.Context, _ *ExecutionContext) error {
	s.ran = true
	return s.err
}

func TestPipeline_Execute(t *testing.T) {
	s1 := &mockStage{name: "stage-1"}
	s2 := &mockStage{name: "stage-2"}
	s3 := &mockStage{name: "stage-3"}

	p := New(s1, s2, s3)
	ec := &ExecutionContext{}

	if err := p.Execute(context.Background(), ec); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	for _, s := range []*mockStage{s1, s2, s3} {
		if !s.ran {
			t.Errorf("stage %q was not executed", s.name)
		}
	}
}

func TestPipeline_Execute_StopsOnError(t *testing.T) {
	s1 := &mockStage{name: "stage-1"}
	s2 := &mockStage{name: "stage-2", err: fmt.Errorf("boom")}
	s3 := &mockStage{name: "stage-3"}

	p := New(s1, s2, s3)
	ec := &ExecutionContext{}

	err := p.Execute(context.Background(), ec)
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}

	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("error should contain 'boom', got: %v", err)
	}
	if !strings.Contains(err.Error(), "stage-2") {
		t.Errorf("error should contain stage name 'stage-2', got: %v", err)
	}

	if !s1.ran {
		t.Error("stage-1 should have run")
	}
	if !s2.ran {
		t.Error("stage-2 should have run (and failed)")
	}
	if s3.ran {
		t.Error("stage-3 should NOT have run after stage-2 failed")
	}
}

func TestPipeline_EmptyStages(t *testing.T) {
	p := New()
	ec := &ExecutionContext{}

	if err := p.Execute(context.Background(), ec); err != nil {
		t.Fatalf("Execute() with empty stages should not error: %v", err)
	}
}

func TestPipeline_Stages(t *testing.T) {
	s1 := &mockStage{name: "a"}
	s2 := &mockStage{name: "b"}

	p := New(s1, s2)
	stages := p.Stages()

	if len(stages) != 2 {
		t.Fatalf("Stages() returned %d, want 2", len(stages))
	}
	if stages[0].Name() != "a" || stages[1].Name() != "b" {
		t.Errorf("Stages() returned wrong stages")
	}
}
