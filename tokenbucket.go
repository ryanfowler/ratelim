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
