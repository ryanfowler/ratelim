package ratelim

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestCreateTBucketQ(t *testing.T) {
	tb1 := NewTBucketQ(-10, time.Second, -10)
	if tb1.burst != 1 {
		t.Error("Burst value should be 1 when a value less than 1 provided")
	}
	if tb1.maxq != 0 {
		t.Error("MaxQ value should be 0 when a negative value is provided")
	}
}

func TestTBucketQGetTok(t *testing.T) {
	tb := NewTBucketQ(10, time.Millisecond*200, 5)
	ch := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func() {
			ch <- tb.GetTok()
		}()
	}
	var val bool
	var negs int
	for i := 0; i < 15; i++ {
		val = <-ch
		if !val {
			negs += 1
		}
	}
	if negs != 5 {
		t.Error("Incorrect amount of tokens granted")
	}
	if tb.toks != 0 {
		t.Error("All tokens should be consumed")
	}
	if tb.qcnt != 5 {
		t.Error("Queue count should be full")
	}
}

func TestTBucketQGetTok2(t *testing.T) {
	tb := NewTBucketQ(10, time.Millisecond*100, 5)
	ch := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func() {
			ch <- tb.GetTok()
		}()
	}
	t1 := time.Now()
	for i := 0; i < 20; i++ {
		<-ch
	}
	dur := time.Now().Sub(t1)
	if dur < time.Millisecond*500 || dur > time.Millisecond*550 {
		t.Error("Incorrect timing of tokens")
	}
	time.Sleep(time.Millisecond * 1200)
	if atomic.LoadInt64(&tb.toks) != 10 {
		t.Error("Token bucket shoudl be full at this point")
	}
}
