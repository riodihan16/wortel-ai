package app

import (
	"context"
	"sync"
	"testing"
	"time"

	"companion/internal/ai"
)

// eventSink collects emitted ChatEvents. The service calls the emitter from
// its streaming goroutine, so every append and read must hold the mutex;
// without this the race detector flags test code, not production code.
type eventSink struct {
	mu     sync.Mutex
	events []ai.ChatEvent
}

func (s *eventSink) emit(_ string, e ai.ChatEvent) {
	s.mu.Lock()
	s.events = append(s.events, e)
	s.mu.Unlock()
}

func (s *eventSink) terminal() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, e := range s.events {
		if e.Type == "cancelled" || e.Type == "completed" {
			n++
		}
	}
	return n
}

func (s *eventSink) snapshot() []ai.ChatEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ai.ChatEvent(nil), s.events...)
}

func TestSendAndCancel(t *testing.T) {
	var sink eventSink
	svc := NewService(&ai.MockProvider{Delay: 20 * time.Millisecond}, sink.emit)
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	reqID, err := svc.SendMessage(ctx, "c1", "halo")
	if err != nil || reqID == "" {
		t.Fatalf("send err: %v id: %q", err, reqID)
	}
	// duplicate while active must fail
	if _, err := svc.SendMessage(ctx, "c1", "lagi"); err == nil {
		t.Fatal("expected duplicate rejection")
	}
	svc.CancelGeneration("c1")
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sink.terminal() > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no terminal event: %+v", sink.snapshot())
}

func TestEmptyRejected(t *testing.T) {
	svc := NewService(&ai.MockProvider{}, func(string, ai.ChatEvent) {})
	if _, err := svc.SendMessage(context.Background(), "c1", "   "); err == nil {
		t.Fatal("expected empty rejection")
	}
}

func TestCancelAllStopsStreams(t *testing.T) {
	var sink eventSink
	svc := NewService(&ai.MockProvider{Delay: 50 * time.Millisecond}, sink.emit)
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	for _, c := range []string{"c1", "c2"} {
		if _, err := svc.SendMessage(ctx, c, "halo"); err != nil {
			t.Fatalf("send %s: %v", c, err)
		}
	}
	svc.CancelAll()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sink.terminal() >= 2 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("expected 2 terminal events: %+v", sink.snapshot())
}

// TestCancelAllWaitsForGoroutines proves CancelAll blocks until every
// generation goroutine has emitted its terminal event, so shutdown cannot
// race a still-running DB write (FR-011). A slow provider that streams well
// past the cancel point is used; if CancelAll returned early the assertion
// on terminal events right after it would fail.
func TestCancelAllWaitsForGoroutines(t *testing.T) {
	var mu sync.Mutex
	var events []ai.ChatEvent
	svc := NewService(&ai.MockProvider{Delay: 10 * time.Millisecond}, func(_ string, e ai.ChatEvent) {
		mu.Lock()
		events = append(events, e)
		mu.Unlock()
	})
	for _, c := range []string{"c1", "c2", "c3"} {
		if _, err := svc.SendMessage(context.Background(), c, "halo"); err != nil {
			t.Fatalf("send %s: %v", c, err)
		}
	}
	svc.CancelAll() // must not return until all goroutines are done

	mu.Lock()
	defer mu.Unlock()
	terminal := 0
	for _, e := range events {
		if e.Type == "cancelled" || e.Type == "completed" {
			terminal++
		}
	}
	if terminal != 3 {
		t.Fatalf("CancelAll returned before goroutines finished: %d/3 terminal: %+v", terminal, events)
	}
	svc.mu.Lock()
	defer svc.mu.Unlock()
	if len(svc.active) != 0 || len(svc.requests) != 0 {
		t.Fatalf("state leaked after CancelAll: active=%d requests=%d", len(svc.active), len(svc.requests))
	}
}
