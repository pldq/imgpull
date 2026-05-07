package utils

import (
	"errors"
	"testing"
	"time"
)

func TestRetry_Do_Success(t *testing.T) {
	retry := NewRetry[int](3)
	attempts := 0

	result, err := retry.Do(func() (int, error) {
		attempts++
		if attempts < 2 {
			return 0, errors.New("not yet")
		}
		return 42, nil
	})

	if err != nil {
		t.Errorf("Do() should succeed, got error: %v", err)
	}
	if result != 42 {
		t.Errorf("Do() should return 42, got %d", result)
	}
	if attempts != 2 {
		t.Errorf("should succeed on 2nd attempt, got %d attempts", attempts)
	}
}

func TestRetry_Do_AllFail(t *testing.T) {
	retry := NewRetry[int](3)
	attempts := 0
	testErr := errors.New("always fail")

	result, err := retry.Do(func() (int, error) {
		attempts++
		return 0, testErr
	})

	if err != testErr {
		t.Errorf("Do() should return last error, got: %v", err)
	}
	if result != 0 {
		t.Errorf("Do() should return zero value on failure, got %d", result)
	}
	if attempts != 3 {
		t.Errorf("should attempt 3 times, got %d attempts", attempts)
	}
}

func TestRetry_Do_FirstSuccess(t *testing.T) {
	retry := NewRetry[int](3)
	attempts := 0

	result, err := retry.Do(func() (int, error) {
		attempts++
		return 42, nil
	})

	if err != nil {
		t.Errorf("Do() should succeed immediately, got error: %v", err)
	}
	if result != 42 {
		t.Errorf("Do() should return 42, got %d", result)
	}
	if attempts != 1 {
		t.Errorf("should succeed on 1st attempt, got %d attempts", attempts)
	}
}

func TestRetry_Do_ZeroAttempts(t *testing.T) {
	retry := NewRetry[int](0)

	result, err := retry.Do(func() (int, error) {
		return 42, nil
	})

	if err == nil {
		t.Error("Do() should error with zero maxAttempts")
	}
	if result != 0 {
		t.Errorf("Do() should return zero value on error, got %d", result)
	}
}

func TestRetry_BackoffTiming(t *testing.T) {
	retry := NewRetry[int](4)
	attempts := 0
	start := time.Now()

	retry.Do(func() (int, error) {
		attempts++
		return 0, errors.New("fail")
	})

	elapsed := time.Since(start)
	// Expected delays: 500ms + 1s + 2s = 3.5s (after 1st, 2nd, 3rd failures)
	// Allow some tolerance
	if elapsed < 3*time.Second || elapsed > 4*time.Second {
		t.Errorf("elapsed time should be around 3.5s, got %v", elapsed)
	}
}
