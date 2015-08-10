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

type TBucketQ struct {
	// number of tokens in the bucket (i.e. available tokens)
	tokens int64
	// bsize is the maximum number of tokens that can fit in the bucket
	bsize int64
	// burst is the number of tokens to add to the bucket each 'tick'
	burst int64
	// qch is the channel over which requests are queued
	qch chan struct{}
	// maxq is the maximum size of the request queue
	maxq int64
	// qcnt is the number of requests waiting in the queue
	qcnt int64
	// ticker contains the channel for adding tokens to the bucket
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

// return a new token bucket with the specified maximum bucket size (burst),
// the time interval for each new token, and the maximum size of the queue
func NewTBucketQ(bsize int64, dur time.Duration, maxq int64) *TBucketQ {
	return NewBurstyTBucketQ(bsize, 1, dur, maxq)
}

func NewBurstyTBucketQ(bsize, burst int64, dur time.Duration, maxq int64) *TBucketQ {
	// verify burst and maxq are acceptable values
	if bsize < 1 {
		bsize = 1
	}
	if burst < 1 {
		burst = 1
	}
	if maxq < 0 {
		maxq = 0
	}
	// create token bucket queue
	tbq := &TBucketQ{
		tokens: bsize,
		bsize:  bsize,
		burst:  burst,
		qch:    make(chan struct{}, 1),
		maxq:   maxq,
		ticker: time.NewTicker(dur),
		cch:    make(chan struct{}, 1),
		prch:   make(chan struct{}, 1),
	}
	// receive from ticker
	go tbq.tick()
	return tbq
}

// grant tokens to requests waiting in the queue or add a token to the bucket
// each time the ticker goes off
func (tbq *TBucketQ) tick() {
	for {
		select {
		case <-tbq.cch:
			// close event received, stop timer and return
			tbq.ticker.Stop()
			return
		case <-tbq.prch:
			// pause event received, listen for stop or resume events
			select {
			case <-tbq.cch:
				// close event received, stop timer and return
				tbq.ticker.Stop()
				return
			case <-tbq.prch:
				// resume event received
			}
		case <-tbq.ticker.C:
			// add token to queue channel if there are any requests waiting
			burst := tbq.burst
			if atomic.LoadInt64(&tbq.qcnt) > 0 {
				for i := int64(0); i < tbq.burst; i++ {
					if qcnt := atomic.LoadInt64(&tbq.qcnt); qcnt > 0 {
						tbq.qch <- struct{}{}
						atomic.AddInt64(&tbq.qcnt, -1)
						burst -= 1
					} else {
						continue
					}
				}
			}
			if burst == 0 {
				continue
			}
			// no requests remaining in queue, attempt to add
			// token(s) to the bucket
			var done bool
			for !done {
				// add token(s) to the bucket if not already full
				if toks := atomic.LoadInt64(&tbq.tokens); toks < tbq.bsize {
					if toks+burst >= tbq.bsize {
						done = atomic.CompareAndSwapInt64(&tbq.tokens, toks, tbq.bsize)
					} else {
						done = atomic.CompareAndSwapInt64(&tbq.tokens, toks, toks+burst)
					}
				} else {
					// bucket is full, throw token(s) away
					done = true
				}
			}
		}
	}
}

// Close stops the internal ticker that adds tokens. The TBucketQ instance is now
// permanently closed and cannpt be reopened. When the TBucketQ will no longer be
// used, this function must be called to stop the internal timer from continuing
// to fire.
//
// It returns true if the TBucketQ has been closed, or false if the TBucketQ has
// already been closed.
func (tbq *TBucketQ) Close() bool {
	if !atomic.CompareAndSwapUint32(&tbq.closed, 0, 1) {
		return false
	}
	tbq.cch <- struct{}{}
	return true
}

// request a token; returns true if token obtained, false otherwise
func (tbq *TBucketQ) GetTok() bool {
	// attempt to obtain token from bucket
	for {
		if toks := atomic.LoadInt64(&tbq.tokens); toks > 0 {
			if atomic.CompareAndSwapInt64(&tbq.tokens, toks, toks-1) {
				return true
			}
			continue
		}
		break
	}
	// no tokens in the bucket, attempt to get on the queue
	var done bool
	for !done {
		if qcnt := atomic.LoadInt64(&tbq.qcnt); qcnt < tbq.maxq {
			done = atomic.CompareAndSwapInt64(&tbq.qcnt, qcnt, qcnt+1)
		} else {
			// queue is full, return false
			return false
		}
	}
	// on queue, wait until token received
	<-tbq.qch
	return true
}

func (tbq *TBucketQ) GetTokNow() bool {
	// attempt to obtain token from bucket
	var done bool
	for !done {
		if toks := atomic.LoadInt64(&tbq.tokens); toks > 0 {
			done = atomic.CompareAndSwapInt64(&tbq.tokens, toks, toks-1)
		} else {
			return false
		}
	}
	return true
}

/*
func (tbq *TBucketQ) GetToks(n int64) bool {
	// attempt to obtain token from bucket
	for {
		if toks := atomic.LoadInt64(&tbq.tokens); toks >= n {
			if atomic.CompareAndSwapInt64(&tbq.tokens, toks, toks-n) {
				return true
			}
			continue
		}
		break
	}
	// no tokens in the bucket, attempt to get on the queue
	var done bool
	for !done {
		if qcnt := atomic.LoadInt64(&tbq.qcnt); qcnt+n <= tbq.maxq {
			done = atomic.CompareAndSwapInt64(&tbq.qcnt, qcnt, qcnt+n)
		} else {
			// queue is full, return false
			return false
		}
	}
	// on queue, wait until token received
	for i := int64(0); i < n; i++ {
		<-tbq.qch
	}
	return true
}
*/

func (tbq *TBucketQ) GetToksNow(n int64) bool {
	// if the provided value is less than 1, return false
	if n < 1 {
		return false
	}
	// attempt to obtain token from bucket
	var done bool
	for !done {
		if toks := atomic.LoadInt64(&tbq.tokens); toks >= n {
			done = atomic.CompareAndSwapInt64(&tbq.tokens, toks, toks-n)
		} else {
			return false
		}
	}
	return true
}

// IsClosed returns true if the TBucketQ has been closed. It returns false if
// it is still open.
func (tbq *TBucketQ) IsClosed() bool {
	if atomic.LoadUint32(&tbq.closed) == 0 {
		return false
	}
	return true
}

// IsPaused returns true if the TBucketQ has been paused. It returns false if
// it is not in a paused state.
func (tbq *TBucketQ) IsPaused() bool {
	if atomic.LoadUint32(&tbq.paused) == 0 {
		return false
	}
	return true
}

// Pause temporarily pauses the TBucketQ from adding new tokens to the bucket.
// When paused, tokens in the bucket can still be retrieved with GetTok or
// GetToks. The TBucketQ can be closed (with Close) or can be resumed (with
// Resume).
//
// This function should be used when the TBucketQ should only be temporarily
// paused. If the TBucketQ will not be used again, Close should be called to
// stop the internal timer.
//
// Pause returns true if the TBucketQ has been paused, or false if the TBucketQ
// is already in a paused state.
func (tbq *TBucketQ) Pause() bool {
	if tbq.IsClosed() || !atomic.CompareAndSwapUint32(&tbq.paused, 0, 1) {
		return false
	}
	tbq.prch <- struct{}{}
	return true
}

// Resume resumes a TBucketQ in a paused state and begins adding new tokens to
// the bucket again.
//
// Resume returns true if the TBucketQ has been resumed, or false if the TBucketQ
// is not in a paused state.
func (tbq *TBucketQ) Resume() bool {
	if tbq.IsClosed() || !atomic.CompareAndSwapUint32(&tbq.paused, 1, 0) {
		return false
	}
	tbq.prch <- struct{}{}
	return true
}
