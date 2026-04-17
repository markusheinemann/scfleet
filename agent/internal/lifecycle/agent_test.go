package lifecycle_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/markusheinemann/scfleet/agent/internal/lifecycle"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

type mockClient struct {
	registerErr  error
	heartbeatErr error
	registered   chan struct{}
	heartbeated  chan struct{}
}

func newMock() *mockClient {
	return &mockClient{
		registered:  make(chan struct{}, 1),
		heartbeated: make(chan struct{}, 100),
	}
}

func (m *mockClient) Register(_ context.Context) error {
	select {
	case m.registered <- struct{}{}:
	default:
	}
	return m.registerErr
}

func (m *mockClient) Heartbeat(_ context.Context) error {
	m.heartbeated <- struct{}{}
	return m.heartbeatErr
}

// waitFor blocks until n signals arrive on ch or the test times out.
func waitFor(t *testing.T, ch <-chan struct{}, n int) {
	t.Helper()
	for range n {
		select {
		case <-ch:
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for signal %d/%d", n, n)
		}
	}
}

func TestRun_CallsRegisterOnce(t *testing.T) {
	mock := newMock()
	agent := lifecycle.New(mock, time.Hour, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- agent.Run(ctx) }()

	waitFor(t, mock.registered, 1)
	// Wait for the immediate heartbeat before cancelling so we know
	// the goroutine has progressed past Register.
	waitFor(t, mock.heartbeated, 1)
	cancel()
	<-done

	// registered channel had capacity 1 and we drained it once — Register was called exactly once.
	select {
	case <-mock.registered:
		t.Error("Register was called more than once")
	default:
	}
}

func TestRun_SendsImmediateHeartbeatAfterRegistration(t *testing.T) {
	mock := newMock()
	// Use a very long ticker interval so no tick fires during the test.
	agent := lifecycle.New(mock, time.Hour, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- agent.Run(ctx) }()

	// Wait for the immediate heartbeat, then cancel before any tick fires.
	waitFor(t, mock.heartbeated, 1)
	cancel()
	<-done

	// Drain any further heartbeats that might have raced in.
	extra := 0
	for {
		select {
		case <-mock.heartbeated:
			extra++
		default:
			goto done
		}
	}
done:
	if extra > 0 {
		t.Errorf("expected only the immediate heartbeat, but got %d extra", extra)
	}
}

func TestRun_CallsHeartbeatOnTick(t *testing.T) {
	mock := newMock()
	agent := lifecycle.New(mock, time.Millisecond, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- agent.Run(ctx) }()

	// 1 immediate + 2 from ticks = 3 total
	waitFor(t, mock.heartbeated, 3)
	cancel()
	<-done
}

func TestRun_ReturnsErrorWhenRegisterFails(t *testing.T) {
	mock := newMock()
	mock.registerErr = errors.New("unauthorized")
	agent := lifecycle.New(mock, time.Millisecond, discardLogger)

	err := agent.Run(context.Background())

	if err == nil {
		t.Fatal("expected error when register fails, got nil")
	}
	select {
	case <-mock.heartbeated:
		t.Error("expected no heartbeats after register failure")
	default:
	}
}

func TestRun_ExitsOnContextCancellation(t *testing.T) {
	mock := newMock()
	agent := lifecycle.New(mock, time.Hour, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- agent.Run(ctx) }()

	waitFor(t, mock.heartbeated, 1)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil on clean shutdown, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

func TestRun_HeartbeatErrorDoesNotStop(t *testing.T) {
	mock := newMock()
	mock.heartbeatErr = errors.New("timeout")
	agent := lifecycle.New(mock, time.Millisecond, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- agent.Run(ctx) }()

	// Loop continues despite errors — wait for several heartbeats.
	waitFor(t, mock.heartbeated, 3)
	cancel()

	if err := <-done; err != nil {
		t.Errorf("expected nil error (heartbeat errors are non-fatal), got %v", err)
	}
}

func TestRun_ReturnsErrorOnZeroInterval(t *testing.T) {
	mock := newMock()
	agent := lifecycle.New(mock, 0, discardLogger)

	err := agent.Run(context.Background())

	if err == nil {
		t.Fatal("expected error for zero interval, got nil")
	}
}

func TestRun_ReturnsErrorOnNegativeInterval(t *testing.T) {
	mock := newMock()
	agent := lifecycle.New(mock, -time.Second, discardLogger)

	err := agent.Run(context.Background())

	if err == nil {
		t.Fatal("expected error for negative interval, got nil")
	}
}
