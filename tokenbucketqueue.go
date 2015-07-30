package ratelim

import (
	"sync/atomic"
	"time"
)

type TBucketQ struct {
	// burst is the maximum number of tokens that can be in the bucket
	burst int64
	// qch is the channel over which requests are queued
	qch chan struct{}
	// maxq is the maximum size of the request queue
	maxq int64
	// qcnt is the number of requests waiting in the queue
	qcnt int64
	// toks is the number of tokens available
	toks int64
	// ticker contains the channel for adding tokens to the bucket
	ticker *time.Ticker
}

// return a new token bucket with the specified maximum bucket size (burst),
// the time interval for each new token, and the maximum size of the queue
func NewTBucketQ(burst int64, dur time.Duration, maxq int64) *TBucketQ {
	// verify burst and maxq are acceptable values
	if burst <= 0 {
		burst = 1
	}
	if maxq < 0 {
		maxq = 0
	}
	// create token bucket queue
	tbq := &TBucketQ{
		burst:  burst,
		qch:    make(chan struct{}, burst),
		maxq:   maxq,
		toks:   burst,
		ticker: time.NewTicker(dur),
	}
	// receive from ticker
	go tbq.tick()
	return tbq
}

// grant tokens to requests waiting in the queue or add a token to the bucket
// each time the ticker goes off
func (tbq *TBucketQ) tick() {
	for {
		<-tbq.ticker.C
		if qcnt := atomic.LoadInt64(&tbq.qcnt); qcnt > 0 {
			tbq.qch <- struct{}{}
			atomic.AddInt64(&tbq.qcnt, -1)
			continue
		}
		var done bool
		for !done {
			toks := atomic.LoadInt64(&tbq.toks)
			if toks < tbq.burst {
				done = atomic.CompareAndSwapInt64(&tbq.toks, toks, toks+1)
			} else {
				done = true
			}
		}
	}
}

// request a token; returns true if token obtained, false otherwise
func (tbq *TBucketQ) GetTok() bool {
	// attempt to obtain token from bucket
	var done bool
	for !done {
		if toks := atomic.LoadInt64(&tbq.toks); toks > 0 {
			done = atomic.CompareAndSwapInt64(&tbq.toks, toks, toks-1)
			if done {
				return true
			}
		} else {
			break
		}
	}
	// attempt to get on the queue
	for !done {
		if qcnt := atomic.LoadInt64(&tbq.qcnt); qcnt < tbq.maxq {
			done = atomic.CompareAndSwapInt64(&tbq.qcnt, qcnt, qcnt+1)
		} else {
			return false
		}
	}
	// on queue, wait until token received
	<-tbq.qch
	return true
}
