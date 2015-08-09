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
	"testing"
	"time"
)

func TestCreateTBucket(t *testing.T) {
	tb := NewTBucket(-10, time.Second)
	if tb.bsize != 1 {
		t.Error("Bucket size value should be 1 when a lower value provided")
	}
	tb2 := NewBurstyTBucket(-10, -15, time.Second)
	if tb2.burst != 1 {
		t.Error("Burst value should be 1 when a lower value provided")
	}
}

func TestTBucketPause(t *testing.T) {
	tb := NewTBucket(10, time.Millisecond)
	if tb.Resume() {
		t.Error("Resume on non-paused TBucket returned true")
	}
	if tb.IsPaused() {
		t.Error("IsPaused returned true on non-paused TBucket")
	}
	if !tb.Pause() {
		t.Error("Pause TBucket failed")
	}
	if !tb.IsPaused() {
		t.Error("IsPaused returned false on paused TBucket")
	}
	tb.Empty()
	if tb.tokens != 0 {
		t.Error("Bucket not emptied")
	}
	time.Sleep(time.Millisecond * 10)
	if tb.tokens != 0 {
		t.Error("Tokens added to paused TBucket")
	}
	if tb.Pause() {
		t.Error("Pause called on paused TBucket returned true")
	}
	if !tb.Resume() {
		t.Error("Resume TBucket failed")
	}
	time.Sleep(time.Millisecond * 10)
	if atomic.LoadInt64(&tb.tokens) == 0 {
		t.Error("Tokens not added to bucket after Resume")
	}
	tb.Pause()
	toks := atomic.LoadInt64(&tb.tokens)
	time.Sleep(time.Millisecond * 5)
	if !tb.Close() {
		t.Error("Close in paused TBucket returned false")
	}
	time.Sleep(time.Millisecond * 5)
	if atomic.LoadInt64(&tb.tokens) != toks {
		t.Error("Tokens added after TBucket closed when paused")
	}
}

func TestTBucketClose(t *testing.T) {
	tb := NewTBucket(10, time.Millisecond)
	if tb.IsClosed() {
		t.Error("IsClosed returned true on open TBucket")
	}
	if !tb.Close() {
		t.Error("Close returned false on open TBucket")
	}
	if !tb.IsClosed() {
		t.Error("IsClosed returned false on closed TBucket")
	}
	tb.Empty()
	time.Sleep(time.Millisecond * 10)
	if atomic.LoadInt64(&tb.tokens) > 0 {
		t.Error("Tokens added to closed TBucket")
	}
	if tb.Close() {
		t.Error("Close returned true on closed TBucket")
	}
}

func TestTBucketFill(t *testing.T) {
	tb := NewTBucket(10, time.Second)
	if !tb.Pause() {
		t.Error("Pause returned false on non-paused TBucket")
	}
	tb.Empty()
	if atomic.LoadInt64(&tb.tokens) != 0 {
		t.Error("Tokens remain in bucket after calling Empty")
	}
	tb.Fill()
	if atomic.LoadInt64(&tb.tokens) != 10 {
		t.Error("Bucket not full after calling Fill:", tb.tokens)
	}
	tb.FillTo(5)
	if atomic.LoadInt64(&tb.tokens) != 5 {
		t.Error("Incorrect number of tokens in bucket after calling FillTo")
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

func TestTBucketGetToks(t *testing.T) {
	tb := NewTBucket(10, time.Second)
	tb.Pause()
	if !tb.GetToks(-10) {
		t.Error("GetToks returned false when enough tokens exist in bucket")
	}
	if atomic.LoadInt64(&tb.tokens) != 9 {
		t.Error("Incorrect number of tokens removed from bucket")
	}
	var oks int
	for i := 0; i < 20; i++ {
		ok := tb.GetToks(1)
		if ok {
			oks += 1
		}
	}
	if oks != 9 {
		t.Error("Incorrect number of tokens provided with GetToks")
	}
}

func BenchmarkGetFail(b *testing.B) {
	tb := NewTBucket(1, time.Millisecond)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.GetTok()
	}
}

func BenchmarkGetSuc(b *testing.B) {
	tb := NewTBucket(10000000000, time.Millisecond)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.GetTok()
	}
}

func BenchmarkGetPFail(b *testing.B) {
	tb := NewTBucket(1, time.Millisecond)
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

func BenchmarkGetPSuc(b *testing.B) {
	tb := NewTBucket(100000000, time.Millisecond)
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
