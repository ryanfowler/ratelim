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
	"sync/atomic"
	"time"
)

type TBucket struct {
	// available tokens
	tokens int64
	// burst is the maximum number of tokens that can be in the bucket
	burst int64
	// ticker is the timer that adds tokens to the bucket
	ticker *time.Ticker
}

// return a new token bucket with the specified maximum bucket size (burst) and
// the time interval for a token to be added to the bucket
func NewTBucket(burst int64, dur time.Duration) *TBucket {
	if burst <= 1 {
		burst = 1
	}
	tb := &TBucket{
		tokens: burst,
		burst:  burst,
		ticker: time.NewTicker(dur),
	}
	go tb.tick()
	return tb
}

// initialize the token bucket
func (tb *TBucket) tick() {
	// receive from ticker
	for {
		<-tb.ticker.C
		var done bool
		for !done {
			val := atomic.LoadInt64(&tb.tokens)
			if val < tb.burst {
				done = atomic.CompareAndSwapInt64(&tb.tokens, val, val+1)
			} else {
				done = true
			}
		}
	}
}

// return true if token obtained; false otherwise
func (tb *TBucket) GetTok() bool {
	var done bool
	var val int64
	for !done {
		val = atomic.LoadInt64(&tb.tokens)
		if val > 0 {
			done = atomic.CompareAndSwapInt64(&tb.tokens, val, val-1)
		} else {
			return false
		}
	}
	return true
}
