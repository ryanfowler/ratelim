package ratelim

import (
	"sync"
	"time"
)

type Limiter struct {
	cache map[string]int64
	max   int64
	mu    sync.Mutex
}

func NewLimiter(max int64, dur time.Duration) *Limiter {
	lim := &Limiter{
		cache: make(map[string]int64),
		mu:    sync.Mutex{},
		max:   max,
	}
	go lim.tick(dur)
	return lim
}

func (lim *Limiter) tick(dur time.Duration) {
	ticker := time.NewTicker(dur)
	for {
		<-ticker.C
		lim.ClearAll()
	}
}

func (lim *Limiter) Inc(key string) bool {
	lim.mu.Lock()
	if lim.cache[key] == lim.max {
		lim.mu.Unlock()
		return false
	}
	lim.cache[key] += 1
	lim.mu.Unlock()
	return true
}

func (lim *Limiter) IncBy(key string, val int64) bool {
	lim.mu.Lock()
	if lim.cache[key]+val > lim.max {
		lim.mu.Unlock()
		return false
	}
	lim.cache[key] += val
	lim.mu.Unlock()
	return true
}

func (lim *Limiter) Dec(key string) bool {
	lim.mu.Lock()
	if lim.cache[key] == 0 {
		lim.mu.Unlock()
		return false
	}
	lim.cache[key] -= 1
	lim.mu.Unlock()
	return true
}

func (lim *Limiter) DecBy(key string, val int64) bool {
	lim.mu.Lock()
	if lim.cache[key]-val < 0 {
		lim.mu.Unlock()
		return false
	}
	lim.cache[key] -= val
	lim.mu.Unlock()
	return true
}

func (lim *Limiter) Clear(key string) {
	lim.mu.Lock()
	delete(lim.cache, key)
	lim.mu.Unlock()
}

func (lim *Limiter) ClearAll() {
	lim.mu.Lock()
	if len(lim.cache) > 0 {
		lim.cache = make(map[string]int64)
	}
	lim.mu.Unlock()
}
