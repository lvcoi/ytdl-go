package downloader

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// parseRateLimit converts human friendly rate strings ("500K", "2M", "1500000",
// "500KiB") into bytes per second.
func parseRateLimit(rate string) (float64, error) {
	rate = strings.TrimSpace(rate)
	if rate == "" {
		return 0, fmt.Errorf("empty rate limit")
	}
	digitsEnd := len(rate)
	for i := 0; i < len(rate); i++ {
		c := rate[i]
		if (c >= '0' && c <= '9') || c == '.' {
			continue
		}
		digitsEnd = i
		break
	}
	if digitsEnd == 0 {
		return 0, fmt.Errorf("invalid rate limit %q", rate)
	}
	numberPart := rate[:digitsEnd]
	suffix := strings.ToLower(strings.TrimSpace(rate[digitsEnd:]))
	var multiplier float64
	switch suffix {
	case "", "b", "bps", "byte", "bytes":
		multiplier = 1
	case "k", "kb":
		multiplier = 1000
	case "kib":
		multiplier = 1024
	case "m", "mb":
		multiplier = 1000 * 1000
	case "mib":
		multiplier = 1024 * 1024
	case "g", "gb":
		multiplier = 1000 * 1000 * 1000
	case "gib":
		multiplier = 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("invalid rate limit suffix %q", suffix)
	}
	value, err := strconv.ParseFloat(numberPart, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid rate limit value %q", numberPart)
	}
	return value * multiplier, nil
}

// rateLimiter implements a token bucket that throttles io.Writer writes to a
// maximum number of bytes per second.
type rateLimiter struct {
	bytesPerSecond float64
	tokens         float64
	last           time.Time
	started        time.Time
	clock          func() time.Time // injectable for tests
	sleep          func(time.Duration) // injectable for tests; advances fake clock
}

func newRateLimiter(bytesPerSecond float64) *rateLimiter {
	if bytesPerSecond <= 0 {
		return nil
	}
	now := time.Now()
	return &rateLimiter{
		bytesPerSecond: bytesPerSecond,
		tokens:         bytesPerSecond,
		last:           now,
		started:        now,
		clock:          time.Now,
	}
}

// wait blocks until at least n tokens are available.
func (r *rateLimiter) wait(ctx context.Context, n int64) error {
	if r == nil {
		return nil
	}
	for {
		now := r.clock()
		elapsed := now.Sub(r.last).Seconds()
		r.last = now
		r.tokens += elapsed * r.bytesPerSecond
		// Allow accumulating beyond the nominal bucket size when a single
		// write is larger than the per-second budget, so large chunks complete.
		cap := r.bytesPerSecond
		if float64(n) > cap {
			cap = float64(n)
		}
		if r.tokens > cap {
			r.tokens = cap
		}
		if r.tokens >= float64(n) {
			r.tokens -= float64(n)
			return nil
		}
		deficit := float64(n) - r.tokens
		sleep := time.Duration(deficit / r.bytesPerSecond * float64(time.Second))
		if sleep > 5*time.Second {
			sleep = 5 * time.Second
		}
		if sleep < 10*time.Millisecond {
			sleep = 10 * time.Millisecond
		}
		if r.sleep != nil {
			r.sleep(sleep)
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}
	}
}

type rateLimitedWriter struct {
	writer  io.Writer
	limiter *rateLimiter
	ctx     context.Context
}

func (w *rateLimitedWriter) Write(p []byte) (int, error) {
	if err := w.limiter.wait(w.ctx, int64(len(p))); err != nil {
		return 0, err
	}
	return w.writer.Write(p)
}

// applyRateLimit wraps a writer with the limiter from opts, if configured.
func applyRateLimit(ctx context.Context, w io.Writer, opts Options) (io.Writer, error) {
	if strings.TrimSpace(opts.LimitRate) == "" {
		return w, nil
	}
	rate, err := parseRateLimit(opts.LimitRate)
	if err != nil {
		return nil, err
	}
	limiter := newRateLimiter(rate)
	if limiter == nil {
		return w, nil
	}
	return &rateLimitedWriter{writer: w, limiter: limiter, ctx: ctx}, nil
}

var autonumberCounter atomic.Int64

// nextAutonumber returns the next per-process sequential download number.
func nextAutonumber() int {
	return int(autonumberCounter.Add(1))
}

// sleepBeforePlaylistEntry sleeps a random duration between SleepInterval and
// MaxSleepInterval (or exactly SleepInterval when no max is given).
func (o *Options) sleepBeforePlaylistEntry(ctx context.Context) error {
	if o.SleepInterval <= 0 {
		return nil
	}
	max := o.MaxSleepInterval
	if max <= 0 || max < o.SleepInterval {
		max = o.SleepInterval
	}
	var d time.Duration
	if max == o.SleepInterval {
		d = o.SleepInterval
	} else {
		d = o.SleepInterval + time.Duration(rand.Int63n(int64(max-o.SleepInterval)))
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// sleepBetweenRequests sleeps for the configured --sleep-requests duration.
func (o *Options) sleepBetweenRequests(ctx context.Context) error {
	if o.SleepRequests <= 0 {
		return nil
	}
	timer := time.NewTimer(o.SleepRequests)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
