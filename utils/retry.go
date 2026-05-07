package utils

import (
	"errors"
	"time"
)

// Retry provides retry functionality with exponential backoff
type Retry[T any] struct {
	maxAttempts int
}

// NewRetry creates a Retry instance
func NewRetry[T any](maxAttempts int) *Retry[T] {
	return &Retry[T]{maxAttempts: maxAttempts}
}

type Callable[T any] func() (T, error)

// Do execute the function with retry logic
// First attempt is immediate, subsequent attempts to have exponential backoff:
// 500ms, 1s, 2s, 4s...
func (r *Retry[T]) Do(fn Callable[T]) (t T, err error) {
	if r.maxAttempts <= 0 {
		return t, errors.New("maxAttempts must be positive")
	}

	var lastErr error
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		t, err = fn()
		if err == nil {
			return t, nil
		}
		lastErr = err

		if attempt < r.maxAttempts {
			delay := time.Duration(500*(1<<(attempt-1))) * time.Millisecond
			time.Sleep(delay)
		}
	}

	return t, lastErr
}
