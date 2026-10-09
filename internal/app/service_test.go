package app

import (
	"context"
	"testing"
	"time"

	"companion/internal/ai"
)

func TestSendAndCancel(t *testing.T) {
	var events []ai.ChatEvent
	svc := NewService(&ai.MockProvider{Delay: 20 * time.Millisecond}, func(_ string, e ai.ChatEvent) {
		events = append(events, e)
	})
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
		done := false
		for _, e := range events {
			if e.Type == "cancelled" || e.Type == "completed" {
				done = true
			}
		}
		if done {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("no terminal event: %+v", events)
}

func TestEmptyRejected(t *testing.T) {
	svc := NewService(&ai.MockProvider{}, func(string, ai.ChatEvent) {})
	if _, err := svc.SendMessage(context.Background(), "c1", "   "); err == nil {
		t.Fatal("expected empty rejection")
	}
}
