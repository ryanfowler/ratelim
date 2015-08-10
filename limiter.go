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
	cache  map[string]int64
	mu     sync.Mutex
	max    int64
	ticker *time.Ticker
	cch    chan struct{}
	closed bool
}

func NewLimiter(max int64, dur time.Duration) *Limiter {
	lim := &Limiter{
		cache:  make(map[string]int64),
		mu:     sync.Mutex{},
		max:    max,
		ticker: time.NewTicker(dur),
		cch:    make(chan struct{}, 1),
	}
	go lim.tick()
	return lim
}

func (lim *Limiter) tick() {
	for {
		select {
		case <-lim.ticker.C:
			lim.ClearAll()
		case <-lim.cch:
			lim.ticker.Stop()
			lim.ClearAll()
			return
		}
	}
}

func (lim *Limiter) Close() {
	lim.mu.Lock()
	if lim.closed {
		lim.mu.Unlock()
		return
	}
	lim.closed = true
	lim.mu.Unlock()
	lim.cch <- struct{}{}
}

func (lim *Limiter) IsClosed() bool {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	return lim.closed
}

func (lim *Limiter) Inc(key string) bool {
	return lim.IncBy(key, 1)
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
	return lim.IncBy(key, -1)
}

func (lim *Limiter) DecBy(key string, val int64) bool {
	return lim.IncBy(key, -val)
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
