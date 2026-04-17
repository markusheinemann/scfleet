package lifecycle_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/markusheinemann/scfleet/agent/internal/lifecycle"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

type mockClient struct {
	registerErr    error
	heartbeatErr   error
	registerCalls  atomic.Int32
	heartbeatCalls atomic.Int32
}

func (m *mockClient) Register(_ context.Context) error {
	m.registerCalls.Add(1)
	return m.registerErr
}

func (m *mockClient) Heartbeat(_ context.Context) error {
	m.heartbeatCalls.Add(1)
	return m.heartbeatErr
}

func TestRun_CallsRegisterOnce(t *testing.T) {
	mock := &mockClient{}
	agent := lifecycle.New(mock, 10*time.Millisecond, discardLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()

	_ = agent.Run(ctx)

	if mock.registerCalls.Load() != 1 {
		t.Errorf("expected Register called once, got %d", mock.registerCalls.Load())
	}
}

func TestRun_SendsImmediateHeartbeatAfterRegistration(t *testing.T) {
	mock := &mockClient{}
	agent := lifecycle.New(mock, time.Hour, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- agent.Run(ctx) }()

	// Give it a moment to register and send the immediate heartbeat, then cancel
	// before any ticker fires.
	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done

	if mock.heartbeatCalls.Load() != 1 {
		t.Errorf("expected 1 immediate heartbeat before first tick, got %d", mock.heartbeatCalls.Load())
	}
}

func TestRun_CallsHeartbeatOnTick(t *testing.T) {
	mock := &mockClient{}
	agent := lifecycle.New(mock, 10*time.Millisecond, discardLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Millisecond)
	defer cancel()

	_ = agent.Run(ctx)

	// 1 immediate + at least 2 from ticks
	if mock.heartbeatCalls.Load() < 3 {
		t.Errorf("expected at least 3 heartbeats (1 immediate + 2 ticks), got %d", mock.heartbeatCalls.Load())
	}
}

func TestRun_ReturnsErrorWhenRegisterFails(t *testing.T) {
	mock := &mockClient{registerErr: errors.New("unauthorized")}
	agent := lifecycle.New(mock, 10*time.Millisecond, discardLogger)

	err := agent.Run(context.Background())

	if err == nil {
		t.Fatal("expected error when register fails, got nil")
	}
	if mock.heartbeatCalls.Load() != 0 {
		t.Errorf("expected no heartbeats after register failure, got %d", mock.heartbeatCalls.Load())
	}
}

func TestRun_ExitsOnContextCancellation(t *testing.T) {
	mock := &mockClient{}
	agent := lifecycle.New(mock, time.Hour, discardLogger)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- agent.Run(ctx) }()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("expected nil on clean shutdown, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

func TestRun_HeartbeatErrorDoesNotStop(t *testing.T) {
	mock := &mockClient{heartbeatErr: errors.New("timeout")}
	agent := lifecycle.New(mock, 10*time.Millisecond, discardLogger)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Millisecond)
	defer cancel()

	err := agent.Run(ctx)

	if err != nil {
		t.Errorf("expected nil error (heartbeat errors are non-fatal), got %v", err)
	}
	if mock.heartbeatCalls.Load() < 2 {
		t.Errorf("expected loop to continue despite heartbeat errors, got %d calls", mock.heartbeatCalls.Load())
	}
}
