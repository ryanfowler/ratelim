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
	"bytes"
	"testing"
	//"time"
)

func TestPoolGet(t *testing.T) {
	p := NewPool(5, func() interface{} {
		return new(bytes.Buffer)
	})
	bufArr := [10]*bytes.Buffer{}
	for i := 0; i < 10; i++ {
		bufArr[i] = p.Get().(*bytes.Buffer)
	}
	for i := 0; i < 10; i++ {
		p.Put(bufArr[i])
	}
	if p.psize != 5 {
		t.Error("Incorrect pool size")
	}
}

/*
func TestPoolGetAndWait(t *testing.T) {
	p := NewPool(5, func() interface{} {
		return new(bytes.Buffer)
	})
	var buf *bytes.Buffer
	var oks int
	ch := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			if item := p.GetAndWait(time.Second); item != nil {
				if _, ok := item.(*bytes.Buffer); !ok {
					t.Error("Cannot perform type cooercion")
				}
				oks += 1
				time.Sleep(time.Millisecond * 50)
				p.Put(buf)
			}
			ch <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-ch
	}
	if oks != 10 {
		t.Error("All items should have been retrieved successfully")
	}
}
*/
