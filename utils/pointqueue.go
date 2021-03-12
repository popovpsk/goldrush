package utils

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type PointQueue struct {
	data [][]point
	m    sync.Mutex
	dgCh chan DigPoint
}

type point struct {
	x, y int32
}

type DigPoint struct {
	X, Y, Amount int32
}

func NewPointQueue() *PointQueue {
	data := make([][]point, 10, 10)
	data[0] = make([]point, 0, 5000)
	data[1] = make([]point, 0, 500)
	data[2] = make([]point, 0, 10)
	for i := 3; i < 10; i++ {
		data[i] = make([]point, 0)
	}

	go func() {
		<-time.After(time.Minute*9 + time.Second*30)
		for i, v := range cntrs {
			if v > 0 {
				fmt.Printf("p %v:%v", i+1, v)
			}
		}
	}()

	return &PointQueue{
		data: data,
		dgCh: make(chan DigPoint),
	}
}

func (dq *PointQueue) Peek() DigPoint {
	dq.m.Lock()
	for i := 9; i >= 0; i-- {
		if len(dq.data[i]) > 0 {
			points := &dq.data[i]
			l := len(*points)
			p := (*points)[l-1]
			*points = (*points)[:l-1]
			dq.m.Unlock()
			return DigPoint{Amount: int32(i + 1), X: p.x, Y: p.y}
		}
	}
	dq.m.Unlock()
	dp := <-dq.dgCh
	return dp
}

var cntrs = []int32{0, 0, 0, 0, 0, 0, 0, 0, 0}

func (dq *PointQueue) Push(p DigPoint) {
	atomic.AddInt32(&cntrs[p.Amount-1], 1)
	select {
	case dq.dgCh <- p:
	default:
		dq.m.Lock()
		dq.data[p.Amount-1] = append(dq.data[p.Amount-1], point{x: p.X, y: p.Y})
		dq.m.Unlock()
	}
}
