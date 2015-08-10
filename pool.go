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

type Pool struct {
	max   int64
	psize int64
	new   func() interface{}
	list  *List
}

func NewPool(max int64, new func() interface{}) *Pool {
	if max < 1 {
		max = 1
	}
	return &Pool{
		max:  max,
		new:  new,
		list: NewList(),
	}
}

func (p *Pool) Empty() {
	for {
		var done bool
		for !done {
			if c := atomic.LoadInt64(&p.psize); c > 0 {
				done = atomic.CompareAndSwapInt64(&p.psize, c, c-1)
			} else {
				// pool is now empty
				return
			}
		}
		p.list.LPop()
	}
}

func (p *Pool) Get() interface{} {
	// attempt to retrieve existing item from pool
	var done bool
	for !done {
		if c := atomic.LoadInt64(&p.psize); c > 0 {
			done = atomic.CompareAndSwapInt64(&p.psize, c, c-1)
		} else {
			// no items in pool, create a new item
			return p.new()
		}
	}
	// return item from pool
	return p.list.LPop()
}

func (p *Pool) Put(item interface{}) {
	// attempt to return item to the pool
	var done bool
	for !done {
		if c := atomic.LoadInt64(&p.psize); c < p.max {
			done = atomic.CompareAndSwapInt64(&p.psize, c, c+1)
		} else {
			// pool is full, discard item
			return
		}
	}
	// add item to pool
	p.list.RPush(item)
}

func (p *Pool) Use(f func(interface{})) {
	v := p.Get()
	f(v)
	p.Put(v)
}
