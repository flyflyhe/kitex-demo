// method_rate_limiter.go
package main

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

// MethodRateLimiter implements limit.RateLimiter
type MethodRateLimiter struct {
	rules    map[string]int64 // key: "Service.Method" → max QPS
	counters sync.Map         // key: string → *methodCounter
}

type methodCounter struct {
	count int64
	reset time.Time
	mu    sync.Mutex
	qps   int64
}

// NewMethodRateLimiter creates a new rate limiter with per-method QPS rules.
func NewMethodRateLimiter(rules map[string]int) *MethodRateLimiter {
	r := make(map[string]int64)
	for k, v := range rules {
		r[k] = int64(v)
	}
	return &MethodRateLimiter{rules: r}
}

// Acquire implements limit.RateLimiter.Acquire
func (m *MethodRateLimiter) Acquire(ctx context.Context) bool {
	ri := rpcinfo.GetRPCInfo(ctx)
	if ri == nil {
		return true // no RPC info, bypass
	}
	key := ri.Invocation().ServiceName() + "." + ri.Invocation().MethodName()

	maxQPS, ok := m.rules[key]
	if !ok {
		return true // no rule, allow
	}

	cnt, _ := m.counters.LoadOrStore(key, &methodCounter{
		qps:   maxQPS,
		reset: time.Now(),
	})
	c := cnt.(*methodCounter)

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if now.Sub(c.reset) >= time.Second {
		atomic.StoreInt64(&c.count, 0)
		c.reset = now
	}

	current := atomic.AddInt64(&c.count, 1)
	return current <= c.qps
}

// Status implements limit.RateLimiter.Status
func (m *MethodRateLimiter) Status(ctx context.Context) (max, current int, interval time.Duration) {
	ri := rpcinfo.GetRPCInfo(ctx)
	if ri == nil {
		return 0, 0, 0
	}
	key := ri.Invocation().ServiceName() + "." + ri.Invocation().MethodName()

	maxQPS, ok := m.rules[key]
	if !ok {
		return 0, 0, 0 // no rule
	}

	cnt, exists := m.counters.Load(key)
	if !exists {
		return int(maxQPS), 0, time.Second
	}

	c := cnt.(*methodCounter)
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ensure count is up-to-date (in case reset hasn't happened yet)
	now := time.Now()
	if now.Sub(c.reset) >= time.Second {
		atomic.StoreInt64(&c.count, 0)
		c.reset = now
	}

	return int(c.qps), int(atomic.LoadInt64(&c.count)), time.Second
}
