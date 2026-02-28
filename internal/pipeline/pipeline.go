package pipeline

import (
	"context"
	"fmt"
	"os"
)

// Stage is a single step in the execution pipeline.
type Stage interface {
	// Name returns a human-readable name for logging.
	Name() string
	// Run executes the stage, reading from and writing to the ExecutionContext.
	Run(ctx context.Context, ec *ExecutionContext) error
}

// Pipeline executes a sequence of stages.
type Pipeline struct {
	stages []Stage
}

// New creates a pipeline from the given stages.
func New(stages ...Stage) *Pipeline {
	return &Pipeline{stages: stages}
}

// Execute runs all stages in sequence.
func (p *Pipeline) Execute(ctx context.Context, ec *ExecutionContext) error {
	for _, s := range p.stages {
		fmt.Fprintf(os.Stderr, "[%s]\n", s.Name())
		if err := s.Run(ctx, ec); err != nil {
			return fmt.Errorf("%s: %w", s.Name(), err)
		}
	}
	return nil
}

// Stages returns the list of stages (for testing).
func (p *Pipeline) Stages() []Stage {
	return p.stages
}
