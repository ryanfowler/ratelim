package ratelim

import (
	//"fmt"
	"testing"
	"time"
)

func TestCreateTBucket(t *testing.T) {
	tb := NewTBucket(-10, time.Second)
	if tb.burst != 1 {
		t.Error("Burst value should be 1 when a lower value provided")
	}
}

func TestTBucketGetTok(t *testing.T) {
	tb := NewTBucket(10, time.Second)
	var oks int
	for i := 0; i < 20; i++ {
		ok := tb.GetTok()
		if ok {
			oks += 1
		}
	}
	if oks != 10 {
		t.Error("Incorrect nubmer of tokens provided")
	}
}

func BenchmarkGet(b *testing.B) {
	tb := NewTBucket(10, time.Millisecond)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.GetTok()
	}
}

func BenchmarkGet2(b *testing.B) {
	tb := NewTBucket(10, time.Millisecond)
	go func() {
		for i := 0; i < 100000000; i++ {
			tb.GetTok()
		}
	}()
	go func() {
		for i := 0; i < 100000000; i++ {
			tb.GetTok()
		}
	}()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.GetTok()
	}
}
