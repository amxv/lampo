package preview

import (
	"context"
	"fmt"
	"time"

	"github.com/amxv/lampo/internal/latex"
)

type completedBuild struct {
	result latex.Result
}

// buildLoop debounces bursts of editor writes and cancels superseded builds.
// Only one latexmk invocation is ever active, so auxiliary files are never
// mutated concurrently and the browser only receives finished PDFs.
func (s *server) buildLoop(ctx context.Context, builder latex.Builder, changes <-chan struct{}) {
	completed := make(chan completedBuild, 1)
	var cancel context.CancelFunc
	var running, pending, ready bool
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}
	defer timer.Stop()
	var timerC <-chan time.Time

	start := func() {
		running = true
		buildCtx, stop := context.WithCancel(ctx)
		cancel = stop
		current := s.snapshot()
		current.Status = "building"
		current.Error = ""
		s.publish(current, nil)
		go func() {
			result := builder.Build(buildCtx)
			select {
			case completed <- completedBuild{result: result}:
			case <-ctx.Done():
			}
		}()
	}
	start()

	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			return
		case <-changes:
			pending = true
			ready = false
			if cancel != nil {
				cancel()
			}
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(110 * time.Millisecond)
			timerC = timer.C
			state := s.snapshot()
			state.Status = "building"
			state.Error = ""
			s.publish(state, nil)
		case <-timerC:
			timerC = nil
			pending = false
			ready = true
			if !running {
				ready = false
				start()
			}
		case done := <-completed:
			running = false
			if cancel != nil {
				cancel()
				cancel = nil
			}
			if !pending && !ready {
				s.accept(done.result)
			}
			if ready {
				ready = false
				start()
			}
		}
	}
}

func (s *server) accept(result latex.Result) {
	next := Snapshot{Log: result.Log, DurationMs: result.Duration.Milliseconds(), UpdatedAt: time.Now()}
	if result.Error != "" {
		next.Status = "error"
		next.Error = result.Error
		s.publish(next, nil)
		_, _ = fmt.Fprintf(s.stdout, "  ✗ %s\n", result.Error)
		return
	}
	next.Status = "ready"
	s.publish(next, result.PDF)
	_, _ = fmt.Fprintf(s.stdout, "  ✓ PDF ready (%d ms)\n", next.DurationMs)
}
