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
	"testing"
	"time"
)

func TestLimiterInc(t *testing.T) {
	lim := NewLimiter(10, time.Second)
	var oks int
	for i := 0; i < 15; i++ {
		if lim.Inc("sample1") {
			oks += 1
		}
	}
	if oks != 10 {
		t.Error("Incorrect increment successes")
	}
}

func TestLimiterDec(t *testing.T) {
	lim := NewLimiter(10, time.Second)
	var oks int
	for i := 0; i < 20; i++ {
		if lim.Inc("sample1") {
			oks += 1
		}
		lim.Inc("sample1")
		lim.Dec("sample1")
		lim.DecBy("sample1", 1)
	}
	if oks != 20 {
		t.Error("Didn't decrement limiter properly")
	}
}

func TestLimiterClear(t *testing.T) {
	lim := NewLimiter(10, time.Second)
	lim.Inc("sample1")
	lim.Clear("sample1")
	if _, ok := lim.cache["sample1"]; ok {
		t.Error("Clear did not remove the value")
	}
}

func TestLimiterClose(t *testing.T) {
	lim := NewLimiter(10, time.Second)
	lim.Inc("sample1")
	lim.Close()
	time.Sleep(time.Millisecond * 100)
	if !lim.IsClosed() {
		t.Error("Close did not actually close the limiter")
	}
	lim.Close()
}
