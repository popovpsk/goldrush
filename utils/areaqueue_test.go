package utils

import (
	"goldrush/api"
	"testing"
	"time"
)

func TestAreaQueue(t *testing.T) {
	aq := NewAreaQueue()

	aq.Push(&api.ExploreResponse{Area: api.Area{SizeX: 100, SizeY: 100}, Amount: 11})
	aq.Push(&api.ExploreResponse{Area: api.Area{SizeX: 100, SizeY: 100}, Amount: 1114})
	aq.Push(&api.ExploreResponse{Area: api.Area{SizeX: 100, SizeY: 100}, Amount: 1115})
	aq.Push(&api.ExploreResponse{Area: api.Area{SizeX: 100, SizeY: 100}, Amount: 32})

	for _, a := range []int{1115, 1114, 32, 11} {
		res := aq.Peek()
		if res.Amount != a {
			t.Errorf("Amount:%v expected:= %v", res, a)
		}
	}

	go func() {
		<-time.After(time.Millisecond * 2)
		aq.Push(&api.ExploreResponse{Area: api.Area{SizeX: 100, SizeY: 100}, Amount: 66})
		aq.Push(&api.ExploreResponse{Area: api.Area{SizeX: 100, SizeY: 100}, Amount: 55})
		aq.Push(&api.ExploreResponse{Area: api.Area{SizeX: 100, SizeY: 100}, Amount: 23})
	}()

	if s := aq.Peek(); s.Amount != 66 {
		t.Errorf("Amount:%v expected:= %v", s.Amount, 66)
	}

	<-time.After(time.Millisecond * 4)

	aq.Push(&api.ExploreResponse{Area: api.Area{SizeX: 100, SizeY: 100}, Amount: 11})

	if s := aq.Peek(); s.Amount != 55 {
		t.Errorf("Amount:%v expected:= %v", s.Amount, 55)
	}
	if s := aq.Peek(); s.Amount != 23 {
		t.Errorf("Amount:%v expected:= %v", s.Amount, 23)
	}
	if s := aq.Peek(); s.Amount != 11 {
		t.Errorf("Amount:%v expected:= %v", s.Amount, 11)
	}
}
