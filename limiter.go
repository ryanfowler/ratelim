// The MIT License (MIT)
//
// Copyright (c) 2015 Ryan Fowler
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
