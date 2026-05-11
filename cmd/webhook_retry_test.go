package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jpillora/backoff"
	"github.com/stretchr/testify/require"
)

func TestInitWithRetry_SucceedsFirstAttempt(t *testing.T) {
	calls := 0
	err := initWithRetry(context.Background(), fastBackoff(), func(_ context.Context) error {
		calls++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, calls)
}

func TestInitWithRetry_RetriesTransientErrors(t *testing.T) {
	calls := 0
	err := initWithRetry(context.Background(), fastBackoff(), func(_ context.Context) error {
		calls++
		if calls < 3 {
			return errors.New("transient error")
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, calls)
}

func TestInitWithRetry_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	err := initWithRetry(ctx, fastBackoff(), func(_ context.Context) error {
		calls++
		return errors.New("always fails")
	})
	require.ErrorIs(t, err, context.Canceled)
	require.GreaterOrEqual(t, calls, 1)
}

func TestInitWithRetry_PreCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := initWithRetry(ctx, fastBackoff(), func(_ context.Context) error {
		calls++
		return errors.New("always fails")
	})
	require.ErrorIs(t, err, context.Canceled)
	require.GreaterOrEqual(t, calls, 1)
}

func fastBackoff() *backoff.Backoff {
	return &backoff.Backoff{Min: time.Millisecond, Max: time.Millisecond}
}
