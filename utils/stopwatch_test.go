package utils

import (
	"errors"
	"testing"
	"time"
)

func TestNewStopWatch(t *testing.T) {
	sw := NewStopWatch()
	if sw == nil {
		t.Error("NewStopWatch() returned nil")
		return
	}
	if len(sw.steps) != 0 {
		t.Errorf("NewStopWatch() steps should be empty, got %d", len(sw.steps))
	}
}

func TestStopWatch_Run_Success(t *testing.T) {
	sw := NewStopWatch()
	err := sw.Run("test-step", func() error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Errorf("Run() unexpected error: %v", err)
	}
	if len(sw.steps) != 1 {
		t.Errorf("Run() should add one step, got %d steps", len(sw.steps))
	}
	if sw.steps[0].Name != "test-step" {
		t.Errorf("Step name: got %q, want %q", sw.steps[0].Name, "test-step")
	}
	if sw.steps[0].Duration <= 0 {
		t.Errorf("Step duration should be positive, got %v", sw.steps[0].Duration)
	}
}

func TestStopWatch_Run_Error(t *testing.T) {
	sw := NewStopWatch()
	testErr := errors.New("test error")
	err := sw.Run("error-step", func() error {
		return testErr
	})
	if !errors.Is(err, testErr) {
		t.Errorf("Run() should return the function error, got %v", err)
	}
	if len(sw.steps) != 1 {
		t.Errorf("Run() should record step even on error, got %d steps", len(sw.steps))
	}
}

func TestStopWatch_Total(t *testing.T) {
	sw := NewStopWatch()
	time.Sleep(10 * time.Millisecond)
	total := sw.Total()
	if total < 10*time.Millisecond {
		t.Errorf("Total() should be at least 10ms, got %v", total)
	}
}

func TestStopWatch_Steps(t *testing.T) {
	sw := NewStopWatch()
	sw.Run("step1", func() error { return nil })
	sw.Run("step2", func() error { return nil })

	steps := sw.Steps()
	if len(steps) != 2 {
		t.Errorf("Steps() should return 2 steps, got %d", len(steps))
	}
	if steps[0].Name != "step1" {
		t.Errorf("First step name: got %q, want %q", steps[0].Name, "step1")
	}
	if steps[1].Name != "step2" {
		t.Errorf("Second step name: got %q, want %q", steps[1].Name, "step2")
	}
}
