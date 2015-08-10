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
)

type Counter struct {
	cnt int64
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) Inc() int64 {
	return atomic.AddInt64(&c.cnt, 1)
}

func (c *Counter) IncBy(i int64) int64 {
	return atomic.AddInt64(&c.cnt, i)
}

func (c *Counter) Dec() int64 {
	return atomic.AddInt64(&c.cnt, -1)
}

func (c *Counter) DecBy(i int64) int64 {
	return atomic.AddInt64(&c.cnt, -i)
}

func (c *Counter) Zero() int64 {
	return atomic.SwapInt64(&c.cnt, 0)
}

func (c *Counter) SetTo(i int64) int64 {
	return atomic.SwapInt64(&c.cnt, i)
}
