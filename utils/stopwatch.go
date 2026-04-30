package utils

import (
	"fmt"
	"time"
)

// Step represents a single step with its name and duration
type Step struct {
	Name     string
	Duration time.Duration
}

// StopWatch records execution time of multiple steps
type StopWatch struct {
	startTime time.Time
	steps     []Step
}

// NewStopWatch creates a new StopWatch
func NewStopWatch() *StopWatch {
	return &StopWatch{
		startTime: time.Now(),
		steps:     make([]Step, 0),
	}
}

// StartTime returns the start time
func (sw *StopWatch) StartTime() time.Time {
	return sw.startTime
}

// Run executes a step function and records its duration
func (sw *StopWatch) Run(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	sw.steps = append(sw.steps, Step{
		Name:     name,
		Duration: duration,
	})

	return err
}

// Total returns the total duration since StopWatch was created
func (sw *StopWatch) Total() time.Duration {
	return time.Since(sw.startTime)
}

// Steps returns all recorded steps
func (sw *StopWatch) Steps() []Step {
	return sw.steps
}

// Print prints all step durations
func (sw *StopWatch) Print() {
	for _, step := range sw.steps {
		fmt.Printf("  %s: %s\n", step.Name, step.Duration.Round(time.Millisecond))
	}
}
