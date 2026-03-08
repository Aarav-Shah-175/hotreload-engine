package engine

import (
	"context"
	"time"
)

const maxBatchBuffer = 64

type Batch struct {
	Count int
	Paths []string
}

type EventBatcher struct {
	notify chan string
	out    chan Batch
}

func NewEventBatcher(ctx context.Context, wait time.Duration) *EventBatcher {
	b := &EventBatcher{
		notify: make(chan string, 128),
		out:    make(chan Batch, 1),
	}

	go func() {
		var timer *time.Timer
		var timerC <-chan time.Time
		buffer := make([]string, 0, maxBatchBuffer)
		total := 0

		flush := func() {
			if total == 0 {
				return
			}
			batch := Batch{Count: total, Paths: append([]string(nil), buffer...)}
			select {
			case b.out <- batch:
			default:
			}
			buffer = buffer[:0]
			total = 0
		}

		for {
			select {
			case <-ctx.Done():
				if timer != nil {
					timer.Stop()
				}
				flush()
				close(b.out)
				return

			case path := <-b.notify:
				total++
				if len(buffer) < maxBatchBuffer {
					buffer = append(buffer, path)
				}
				if timer == nil {
					timer = time.NewTimer(wait)
					timerC = timer.C
					continue
				}
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(wait)

			case <-timerC:
				flush()
			}
		}
	}()

	return b
}

func (b *EventBatcher) Add(path string) {
	select {
	case b.notify <- path:
	default:
	}
}

func (b *EventBatcher) C() <-chan Batch {
	return b.out
}
