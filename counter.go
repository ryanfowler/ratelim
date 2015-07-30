package ratelim

import (
	"sync/atomic"
)

type Counter struct {
	cnt *int64
}

func NewCounter() *Counter {
	var cnt int64
	return &Counter{cnt: &cnt}
}

func (c *Counter) Inc() int64 {
	return atomic.AddInt64(c.cnt, 1)
}

func (c *Counter) IncBy(i int64) int64 {
	return atomic.AddInt64(c.cnt, i)
}

func (c *Counter) Dec() int64 {
	return atomic.AddInt64(c.cnt, -1)
}

func (c *Counter) DecBy(i int64) int64 {
	return atomic.AddInt64(c.cnt, -i)
}

func (c *Counter) Zero() int64 {
	return atomic.SwapInt64(c.cnt, 0)
}

func (c *Counter) SetTo(i int64) int64 {
	return atomic.SwapInt64(c.cnt, i)
}
