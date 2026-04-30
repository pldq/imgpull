package utils

import (
	"errors"
	"time"
)

// Retry provides retry functionality with exponential backoff
type Retry struct {
	maxAttempts int
}

// NewRetry creates a Retry instance
func NewRetry(maxAttempts int) *Retry {
	return &Retry{maxAttempts: maxAttempts}
}

// Do executes the function with retry logic
// First attempt is immediate, subsequent attempts have exponential backoff:
// 500ms, 1s, 2s, 4s...
func (r *Retry) Do(fn func() error) error {
	if r.maxAttempts <= 0 {
		return errors.New("maxAttempts must be positive")
	}

	var lastErr error
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		if attempt < r.maxAttempts {
			delay := time.Duration(500*(1<<(attempt-1))) * time.Millisecond
			time.Sleep(delay)
		}
	}

	return lastErr
}
