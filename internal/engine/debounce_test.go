package engine

import (
	"context"
	"testing"
	"time"
)

func TestDebouncerCoalescesBurst(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := NewDebouncer(ctx, 60*time.Millisecond)
	d.Notify()
	time.Sleep(15 * time.Millisecond)
	d.Notify()
	time.Sleep(15 * time.Millisecond)
	d.Notify()

	select {
	case <-d.C():
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected a debounced trigger")
	}

	select {
	case <-d.C():
		t.Fatal("expected exactly one trigger for burst")
	case <-time.After(120 * time.Millisecond):
	}
}

func TestDebouncerCanTriggerAgain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := NewDebouncer(ctx, 40*time.Millisecond)
	d.Notify()
	<-d.C()
	d.Notify()

	select {
	case <-d.C():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected second debounced trigger")
	}
}
