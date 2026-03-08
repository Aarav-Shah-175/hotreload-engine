package engine

import (
	"context"
	"testing"
	"time"
)

func TestEventBatcherCoalescesRapidEventsIntoSingleBatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batcher := NewEventBatcher(ctx, 60*time.Millisecond)
	for i := 0; i < 7; i++ {
		batcher.Add("file.go")
		time.Sleep(10 * time.Millisecond)
	}

	var batch Batch
	select {
	case batch = <-batcher.C():
	case <-time.After(300 * time.Millisecond):
		t.Fatal("expected a debounced batch")
	}

	if batch.Count != 7 {
		t.Fatalf("batch.Count = %d, want 7", batch.Count)
	}

	select {
	case <-batcher.C():
		t.Fatal("expected only one batch for rapid burst")
	case <-time.After(120 * time.Millisecond):
	}
}

func TestRapidEventsTriggerSingleReloadDecision(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batcher := NewEventBatcher(ctx, 50*time.Millisecond)
	for i := 0; i < 10; i++ {
		batcher.Add("main.go")
	}

	reloads := 0
	timer := time.NewTimer(220 * time.Millisecond)
	defer timer.Stop()

loop:
	for {
		select {
		case <-batcher.C():
			reloads++
		case <-timer.C:
			break loop
		}
	}

	if reloads != 1 {
		t.Fatalf("reload decisions = %d, want 1", reloads)
	}
}

func TestEventBatcherCanEmitMultipleBatches(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batcher := NewEventBatcher(ctx, 40*time.Millisecond)
	batcher.Add("a.go")
	<-batcher.C()
	batcher.Add("b.go")

	select {
	case batch := <-batcher.C():
		if batch.Count != 1 {
			t.Fatalf("batch.Count = %d, want 1", batch.Count)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected second batch")
	}
}

func TestEventBatcherBuffersOnlySmallPathSample(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	batcher := NewEventBatcher(ctx, 40*time.Millisecond)
	for i := 0; i < maxBatchBuffer+10; i++ {
		batcher.Add("x.go")
	}

	select {
	case batch := <-batcher.C():
		if batch.Count != maxBatchBuffer+10 {
			t.Fatalf("batch.Count = %d, want %d", batch.Count, maxBatchBuffer+10)
		}
		if len(batch.Paths) != maxBatchBuffer {
			t.Fatalf("len(batch.Paths) = %d, want %d", len(batch.Paths), maxBatchBuffer)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected batch")
	}
}
