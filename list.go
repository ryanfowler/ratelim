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
	"sync"
)

// lnode represents a single node in a List
type lnode struct {
	// next is the next item in the List
	next *lnode
	// prev is the previous item in the List
	prev *lnode
	// value is the actual value for the node
	value interface{}
}

// List represents a light-weight implementation of a doubly linked list.
//
// To learn more about doubly linked lists, visit:
// https://en.wikipedia.org/wiki/Doubly_linked_list
type List struct {
	// head is a pointer to the head node
	head *lnode
	// tail is a pointer to the tail node
	tail *lnode
	// mu is the mutex for accessing any list data
	mu sync.Mutex
	// length is the total length of the list
	length int
}

// NewList returns an initialized List pointer.
func NewList() *List {
	return &List{
		mu: sync.Mutex{},
	}
}

// Empty removes all items from the List. After calling this function, the
// length of the list is 0.
func (l *List) Empty() {
	l.mu.Lock()
	l.head = nil
	l.tail = nil
	l.length = 0
	l.mu.Unlock()
}

// LEach iterates over each item in the List, starting from the left-most node
// (the head).
//
// The parameter "f" is a function that accepts the index of the node
// (zero-based), and the value of the current node. This function should return
// true to continue the iteration, or false to immediately stop and return from
// the LEach function.
func (l *List) LEach(f func(int, interface{}) bool) {
	l.mu.Lock()
	n := l.head
	var c int
	for n != nil {
		if !f(c, n.value) {
			l.mu.Unlock()
			return
		}
		c += 1
		n = n.next
	}
	l.mu.Unlock()
}

// LPop removes the left-most node from the List (i.e. the head), and returns
// it's value.
func (l *List) LPop() interface{} {
	// lock the list, unlock on return
	l.mu.Lock()
	// retrieve head
	if h := l.head; h != nil {
		if l.length == 1 {
			l.length = 0
			l.head = nil
			l.tail = nil
			l.mu.Unlock()
			return h.value
		}
		l.length -= 1
		l.head = h.next
		l.head.prev = nil
		l.mu.Unlock()
		return h.value
	}
	l.mu.Unlock()
	return nil
}

// LPush inserts the provided value to the left-most position in the list (the
// head position).
func (l *List) LPush(v interface{}) {
	// create new node
	n := &lnode{
		value: v,
	}
	// lock the list, unlock on return
	l.mu.Lock()
	// add node to head
	l.length += 1
	if h := l.head; h != nil {
		h.prev = n
		n.next = h
		l.head = n
		l.mu.Unlock()
		return
	}
	// list is empty, assign node to head & tail
	l.head = n
	l.tail = n
	l.mu.Unlock()
}

// REach iterates over each item in the List, starting from the right-most node
// (the tail).
//
// The parameter "f" is a function that accepts the index of the node
// (zero-based), and the value of the current node. This function should return
// true to continue the iteration, or false to immediately stop and return from
// the REach function.
func (l *List) REach(f func(int, interface{}) bool) {
	l.mu.Lock()
	n := l.tail
	c := l.length - 1
	for n != nil {
		if !f(c, n.value) {
			l.mu.Unlock()
			return
		}
		c -= 1
		n = n.prev
	}
	l.mu.Unlock()
}

// RPop removes the right-most node from the List (i.e. the tail), and returns
// it's value.
func (l *List) RPop() interface{} {
	// lock the list, unlock on return
	l.mu.Lock()
	// retrieve tail
	if t := l.tail; t != nil {
		if l.length == 1 {
			l.length = 0
			l.tail = nil
			l.head = nil
			l.mu.Unlock()
			return t.value
		}
		l.length -= 1
		l.tail = t.prev
		l.tail.next = nil
		l.mu.Unlock()
		return t.value
	}
	l.mu.Unlock()
	return nil
}

// RPush inserts the provided value to the right-most position in the list (the
// tail position).
func (l *List) RPush(v interface{}) {
	// create new node
	n := &lnode{
		value: v,
	}
	// lock the list, unlock on return
	l.mu.Lock()
	// add node to tail
	l.length += 1
	if t := l.tail; t != nil {
		n.prev = t
		t.next = n
		l.tail = n
		l.mu.Unlock()
		return
	}
	// list is empty, assign node to head & tail
	l.head = n
	l.tail = n
	l.mu.Unlock()
}

// ValueAt returns the value of the node at position "i". If the provided
// position is not in the bounds of the List, nil is returned.
func (l *List) ValueAt(i int) interface{} {
	l.mu.Lock()
	defer l.mu.Unlock()
	if i >= l.length || i < 0 {
		return nil
	}
	var c int
	n := l.head
	for {
		if c == i {
			return n.value
		}
		n = n.next
		c += 1
	}
}
