package pointqueue

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type PointQueue struct {
	points         []point
	priorityPoints []DigPoint
	m              sync.Mutex
	dgCh           chan DigPoint
}

type point struct {
	x, y int32
}

type DigPoint struct {
	X, Y, Amount int32
}

var cntrs1 int32
var cntrs2 int32

func EndLog() {
	fmt.Printf("PointsCall== Pushed:%v : Peek:%v\n", atomic.LoadInt32(&cntrs1), atomic.LoadInt32(&cntrs2))
}

func NewPointQueue() *PointQueue {
	data := make([]point, 0, 20000)
	priorityPoints := make([]DigPoint, 0, 500)
	return &PointQueue{
		points:         data,
		priorityPoints: priorityPoints,
		dgCh:           make(chan DigPoint),
	}
}

func (dq *PointQueue) Peek() DigPoint {
	atomic.AddInt32(&cntrs2, 1)
	var result DigPoint
	dq.m.Lock()

	if len(dq.priorityPoints) > 0 {
		l := len(dq.priorityPoints)
		result = dq.priorityPoints[l-1]
		dq.priorityPoints = dq.priorityPoints[:l-1]
		dq.m.Unlock()
	} else if len(dq.points) > 0 {
		l := len(dq.points)
		p := dq.points[l-1]
		result.X, result.Y = p.x, p.y
		result.Amount = 1
		dq.points = dq.points[:l-1]
		dq.m.Unlock()
	} else {
		dq.m.Unlock()
		result = <-dq.dgCh
	}
	return result
}

func (dq *PointQueue) Push(p DigPoint) {
	atomic.AddInt32(&cntrs1, 1)
	select {
	case dq.dgCh <- p:
	default:
		dq.m.Lock()
		if p.Amount == 1 {
			dq.points = append(dq.points, point{p.X, p.Y})
		} else {
			dq.priorityPoints = append(dq.priorityPoints, p)
		}
		dq.m.Unlock()
	}
}
