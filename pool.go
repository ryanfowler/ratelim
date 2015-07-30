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

type Pool struct {
	created int64
	max     int
	ch      chan interface{}
	new     func() interface{}
}

func NewPool(max int, new func() interface{}) *Pool {
	return &Pool{
		max: max,
		ch:  make(chan interface{}, max),
		new: new,
	}
}

func (p *Pool) Get() interface{} {
	var item interface{}
	select {
	case item = <-p.ch:
	default:
		var done bool
		for !done {
			c := atomic.LoadInt64(&p.created)
			if c < int64(p.max) {
				done = atomic.CompareAndSwapInt64(&p.created, c, c+1)
			} else {
				return item
			}
		}
		item = p.new()
	}
	return item
}

// doesn't work
func (p *Pool) GetAndWait(dur time.Duration) interface{} {
	var item interface{}
	select {
	case item = <-p.ch:
	default:
		var done bool
		for !done {
			c := atomic.LoadInt64(&p.created)
			if c < int64(p.max) {
				done = atomic.CompareAndSwapInt64(&p.created, c, c+1)
			} else {
				<-time.NewTimer(dur).C
				return item
			}
		}
		item = p.new
	}
	return item
}

func (p *Pool) Put(item interface{}) {
	p.ch <- item
}
