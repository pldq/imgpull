package utils

import (
	"errors"
	"testing"
	"time"
)

func TestRetry_Do_Success(t *testing.T) {
	retry := NewRetry(3)
	attempts := 0

	err := retry.Do(func() error {
		attempts++
		if attempts < 2 {
			return errors.New("not yet")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Do() should succeed, got error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("should succeed on 2nd attempt, got %d attempts", attempts)
	}
}

func TestRetry_Do_AllFail(t *testing.T) {
	retry := NewRetry(3)
	attempts := 0
	testErr := errors.New("always fail")

	err := retry.Do(func() error {
		attempts++
		return testErr
	})

	if err != testErr {
		t.Errorf("Do() should return last error, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("should attempt 3 times, got %d attempts", attempts)
	}
}

func TestRetry_Do_FirstSuccess(t *testing.T) {
	retry := NewRetry(3)
	attempts := 0

	err := retry.Do(func() error {
		attempts++
		return nil
	})

	if err != nil {
		t.Errorf("Do() should succeed immediately, got error: %v", err)
	}
	if attempts != 1 {
		t.Errorf("should succeed on 1st attempt, got %d attempts", attempts)
	}
}

func TestRetry_Do_ZeroAttempts(t *testing.T) {
	retry := NewRetry(0)

	err := retry.Do(func() error {
		return nil
	})

	if err == nil {
		t.Error("Do() should error with zero maxAttempts")
	}
}

func TestRetry_BackoffTiming(t *testing.T) {
	retry := NewRetry(4)
	attempts := 0
	start := time.Now()

	retry.Do(func() error {
		attempts++
		return errors.New("fail")
	})

	elapsed := time.Since(start)
	// Expected delays: 500ms + 1s + 2s = 3.5s (after 1st, 2nd, 3rd failures)
	// Allow some tolerance
	if elapsed < 3*time.Second || elapsed > 4*time.Second {
		t.Errorf("elapsed time should be around 3.5s, got %v", elapsed)
	}
}
