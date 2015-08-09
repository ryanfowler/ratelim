package ratelim

import (
	"sync"
	"testing"
)

var (
	sampleVal  = "TestVal"
	sampleChan = make(chan struct{}, 1)
)

func TestListL(t *testing.T) {
	l := &List{
		mu: sync.Mutex{},
	}
	l.LPush("Val1")
	if l.length != 1 {
		t.Error("LPush length incorrect")
	}
	l.LPush("Val2")
	l.LPush("Val3")
	if l.length != 3 {
		t.Error("LPush length incorrect")
	}
	v3 := l.LPop().(string)
	if v3 != "Val3" {
		t.Error("LPop value incorrect")
	}
	v1 := l.RPop().(string)
	if v1 != "Val1" {
		t.Error("RPop value incorrect")
	}
	if l.length != 1 {
		t.Error("Length incorrect after LPop/RPop")
	}
}

func TestListR(t *testing.T) {
	l := &List{
		mu: sync.Mutex{},
	}
	l.RPush("Val1")
	if l.length != 1 {
		t.Error("Length incorrect after calling RPush")
	}
	l.LPush("Val2")
	l.RPush("Val3")
	v0 := l.ValueAt(0).(string)
	v1 := l.ValueAt(1).(string)
	v2 := l.ValueAt(2).(string)
	if v0 != "Val2" {
		t.Error("Value at 0 incorrect")
	}
	if v1 != "Val1" {
		t.Error("Value at 1 incorrect")
	}
	if v2 != "Val3" {
		t.Error("Value at 2 incorrect")
	}
}

func BenchmarkListLPush(b *testing.B) {
	l := &List{
		mu: sync.Mutex{},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.LPush(sampleVal)
	}
}

func BenchmarkListLPop(b *testing.B) {
	l := &List{
		mu: sync.Mutex{},
	}
	var v string
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.LPush(sampleVal)
		l.LPush(sampleVal)
		v = l.LPop().(string)
	}
	b.StopTimer()
	if v == "hello" {
	}
}
