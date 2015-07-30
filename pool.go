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
