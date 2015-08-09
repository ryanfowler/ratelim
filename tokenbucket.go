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

// TBucket is a struct representing the token bucket algorithm.
//
// All provided functions are safe for concurrent use. It is designed to be a
// fast, lightweight, and lock-free implementation of the token bucket algorithm.
//
// For more information on the token bucket algorithm, check out:
// https://en.wikipedia.org/wiki/Token_bucket
type TBucket struct {
	// available tokens
	tokens int64
	// bsize is the maximum number of tokens that can fit in the bucket
	bsize int64
	// burst is the number of tokens to add to the bucket for each 'tick'
	burst int64
	// ticker is the timer that adds tokens to the bucket
	ticker *time.Ticker
	// cch is the channel that listens for a close event
	cch chan struct{}
	// closed indicates whether the bucket is closed (1) or not (0)
	closed uint32
	// prch is the channel that listens for a pause/resume event
	prch chan struct{}
	// paused indicates whether the bucket is paused (1) or not (0)
	paused uint32
}

// NewTBucket will create a new token bucket instance.
//
// The parameter "bsize" is the maximum bucket size and the parameter "dur" is
// the time interval another token will be added to the bucket. Using this
// function is equivilant to calling NewBurstyTBucket(bsize, 1, dur).
//
// To learn more about the token bucket algorithm, check out:
// https://en.wikipedia.org/wiki/Token_bucket
func NewTBucket(bsize int64, dur time.Duration) *TBucket {
	return NewBurstyTBucket(bsize, 1, dur)
}

// NewBurstyTBucket will create a new token bucket instance.
//
// The parameter "bsize" is the maximum bucket size and the parameter "burst" is
// the number of tokens to be added to the bucket every time interval "dur".
//
// To learn more about the token bucket algorithm, check out:
// https://en.wikipedia.org/wiki/Token_bucket
func NewBurstyTBucket(bsize, burst int64, dur time.Duration) *TBucket {
	if bsize < 1 {
		bsize = 1
	}
	if burst < 1 {
		burst = 1
	}
	tb := &TBucket{
		tokens: bsize,
		bsize:  bsize,
		burst:  burst,
		ticker: time.NewTicker(dur),
		cch:    make(chan struct{}, 1),
		prch:   make(chan struct{}, 1),
	}
	go tb.tick()
	return tb
}

// This function should run in it's own goroutine and adds tokens to the bucket
// as necessary, as well as waits for the close and pause/resume commands.
func (tb *TBucket) tick() {
	for {
		select {
		case <-tb.cch:
			// close event received, stop timer and return
			tb.ticker.Stop()
			return
		case <-tb.prch:
			// pause event received, listen for stop or resume events
			select {
			case <-tb.cch:
				// close event received, stop timer and return
				tb.ticker.Stop()
				return
			case <-tb.prch:
				// resume event received
			}
		case <-tb.ticker.C:
			// timer event, attempt to add token to the bucket
			var done bool
			for !done {
				toks := atomic.LoadInt64(&tb.tokens)
				if toks < tb.bsize {
					done = atomic.CompareAndSwapInt64(&tb.tokens, toks, toks+tb.burst)
				} else {
					done = true
				}
			}
		}
	}
}

// Close stops the internal ticker that adds tokens. The TBucket instance is now
// permanently closed and cannpt be reopened. When the TBucket will no longer be
// used, this function must be called to stop the internal timer from continuing
// to fire.
//
// It returns true if the TBucket has been closed, or false if the TBucket has
// already been closed.
func (tb *TBucket) Close() bool {
	if !atomic.CompareAndSwapUint32(&tb.closed, 0, 1) {
		return false
	}
	tb.cch <- struct{}{}
	return true
}

// Empty removes all tokens from the bucket. This function does not stop the
// timer that adds new tokens to the bucket.
func (tb *TBucket) Empty() {
	atomic.StoreInt64(&tb.tokens, 0)
}

// Fill adds the maximum amount of tokens to the bucket (fills the bucket)
// according to the defined bucket size.
func (tb *TBucket) Fill() {
	atomic.StoreInt64(&tb.tokens, tb.bsize)
}

// FillTo adds "n" tokens to the bucket. The value "n" may be larger than the
// defined bucket size.
func (tb *TBucket) FillTo(n int64) {
	atomic.StoreInt64(&tb.tokens, n)
}

// GetTok attempts to retrieve a single token from the bucket.
//
// It returns true if a token has been successfully retrieved, or returns
// false if no token is available (the bucket is empty).
func (tb *TBucket) GetTok() bool {
	var done bool
	for !done {
		toks := atomic.LoadInt64(&tb.tokens)
		if toks > 0 {
			done = atomic.CompareAndSwapInt64(&tb.tokens, toks, toks-1)
		} else {
			return false
		}
	}
	return true
}

// GetToks attempts to retrieve "n" tokens from the bucket.
//
// It returns true if "n" tokens have been successfully retrieved, or returns
// false if there are not enough tokens available.
//
// The provided parameter "n" cannot be smaller than 1. If a smaller value is
// provided, the value 1 will be used.
func (tb *TBucket) GetToks(n int64) bool {
	if n < 1 {
		n = 1
	}
	var done bool
	for !done {
		toks := atomic.LoadInt64(&tb.tokens)
		if toks >= n {
			done = atomic.CompareAndSwapInt64(&tb.tokens, toks, toks-n)
		} else {
			return false
		}
	}
	return true
}

// IsClosed returns true if the TBucket has been closed. It returns false if
// it is still open.
func (tb *TBucket) IsClosed() bool {
	if atomic.LoadUint32(&tb.closed) == 0 {
		return false
	}
	return true
}

// IsPaused returns true if the TBucket has been paused. It returns false if
// it is not in a paused state.
func (tb *TBucket) IsPaused() bool {
	if atomic.LoadUint32(&tb.paused) == 0 {
		return false
	}
	return true
}

// Pause temporarily pauses the TBucket from adding new tokens to the bucket.
// When paused, tokens in the bucket can still be retrieved with GetTok or
// GetToks. The TBucket can be closed (with Close) or can be resumed (with
// Resume).
//
// This function should be used when the TBucket should only be temporarily
// paused. If the TBucket will not be used again, Close should be called to
// stop the internal timer.
//
// Pause returns true if the TBucket has been paused, or false if the TBucket
// is already in a paused state.
func (tb *TBucket) Pause() bool {
	if tb.IsClosed() || !atomic.CompareAndSwapUint32(&tb.paused, 0, 1) {
		return false
	}
	tb.prch <- struct{}{}
	return true
}

// Resume resumes a TBucket in a paused state and starts adding new tokens to
// the bucket again.
//
// Resume returns true if the TBucket has been resumed, or false if the TBucket
// is not in a paused state.
func (tb *TBucket) Resume() bool {
	if tb.IsClosed() || !atomic.CompareAndSwapUint32(&tb.paused, 1, 0) {
		return false
	}
	tb.prch <- struct{}{}
	return true
}
