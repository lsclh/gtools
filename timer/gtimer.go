package timer

import (
	"context"
	"sync"
	"time"

	"github.com/lsclh/gtools/timer/type"
)

// Timer is the timer manager, which uses ticks to calculate the timing interval.
type Timer struct {
	mu      sync.RWMutex
	queue   *priorityQueue // queue is a priority queue based on heap structure.
	status  *gtype.Int     // status is the current timer status.
	ticks   *gtype.Int64   // ticks is the proceeded interval number by the timer.
	options TimerOptions   // timer options is used for timer configuration.
}

// TimerOptions is the configuration object for Timer.
type TimerOptions struct {
	Interval time.Duration // (optional) Interval is the underlying rolling interval tick of the timer.
	Quick    bool          // Quick is used for quick timer, which means the timer will not wait for the first interval to be elapsed.
}

const (
	StatusReady          = 0      // Job or Timer is ready for running.
	StatusRunning        = 1      // Job or Timer is already running.
	StatusStopped        = 2      // Job or Timer is stopped.
	StatusClosed         = -1     // Job or Timer is closed and waiting to be deleted.
	panicExit            = "exit" // panicExit is used for custom job exit with panic.
	defaultTimerInterval = 100    // defaultTimerInterval is the default timer interval in milliseconds.
)

var (
	defaultInterval = getDefaultInterval()
	defaultTimer    = New()
)

func getDefaultInterval() time.Duration {
	return time.Duration(defaultTimerInterval) * time.Millisecond
}

// DefaultOptions creates and returns a default options object for Timer creation.
func DefaultOptions() TimerOptions {
	return TimerOptions{
		Interval: defaultInterval,
	}
}

// SetTimeout runs the job once after duration of `delay`.
// It is like the one in javascript.
func SetTimeout(ctx context.Context, delay time.Duration, job JobFunc) {
	AddOnce(ctx, delay, job)
}

// SetInterval runs the job every duration of `delay`.
// It is like the one in javascript.
func SetInterval(ctx context.Context, interval time.Duration, job JobFunc) {
	Add(ctx, interval, job)
}

// Add adds a timing job to the default timer, which runs in interval of `interval`.
func Add(ctx context.Context, interval time.Duration, job JobFunc) *Entry {
	return defaultTimer.Add(ctx, interval, job)
}

// AddOnce is a convenience function for adding a job which only runs once and then exits.
func AddOnce(ctx context.Context, interval time.Duration, job JobFunc) *Entry {
	return defaultTimer.AddOnce(ctx, interval, job)
}
