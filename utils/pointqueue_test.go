package utils

import (
	"testing"
	"time"
)

func TestPointQueue(t *testing.T) {
	pq := NewPointQueue()

	go func() {
		a := pq.Peek()
		if a.Amount != 5 {
			t.Errorf("Amount:%v expected:= %v", a.Amount, 5)
		}
	}()

	<-time.After(time.Millisecond)
	pq.Push(DigPoint{X: 222, Y: 111, Amount: 5})

	pq.Push(DigPoint{X: 222, Y: 111, Amount: 2})
	pq.Push(DigPoint{X: 333, Y: 111, Amount: 7})
	pq.Push(DigPoint{X: 111, Y: 111, Amount: 1})
	pq.Push(DigPoint{X: 222, Y: 111, Amount: 2})
	pq.Push(DigPoint{X: 333, Y: 111, Amount: 9})
	pq.Push(DigPoint{X: 333, Y: 111, Amount: 3})

	s := []int32{9, 7, 3, 2, 2, 1, 6}

	go func() {
		<-time.After(time.Millisecond * 10)
		pq.Push(DigPoint{X: 333, Y: 111, Amount: 6})
	}()

	for _, v := range s {
		a := pq.Peek()

		if a.Amount != v {
			t.Errorf("Amount:%v expected:= %v", a.Amount, v)
		}
	}

}
