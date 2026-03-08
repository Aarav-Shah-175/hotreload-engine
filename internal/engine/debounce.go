package engine

import (
	"context"
	"time"
)

type Debouncer struct {
	notify chan struct{}
	out    chan struct{}
}

func NewDebouncer(ctx context.Context, wait time.Duration) *Debouncer {
	d := &Debouncer{
		notify: make(chan struct{}, 1),
		out:    make(chan struct{}, 1),
	}

	go func() {
		var timer *time.Timer
		var timerC <-chan time.Time

		for {
			select {
			case <-ctx.Done():
				if timer != nil {
					timer.Stop()
				}
				close(d.out)
				return
			case <-d.notify:
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
				select {
				case d.out <- struct{}{}:
				default:
				}
			}
		}
	}()

	return d
}

func (d *Debouncer) Notify() {
	select {
	case d.notify <- struct{}{}:
	default:
	}
}

func (d *Debouncer) C() <-chan struct{} {
	return d.out
}
