package utils

import (
	"container/heap"
	"goldrush/api"
	"sync"
)

type AreaQueue struct {
	pq *PriorityQueue
	l  sync.Mutex
	ch chan *api.ExploreResponse
}

func NewAreaQueue() *AreaQueue {
	aq := &AreaQueue{
		ch: make(chan *api.ExploreResponse),
	}
	pq := make(PriorityQueue, 0, 50)
	aq.pq = &pq
	heap.Init(aq.pq)
	return aq
}

func (q *AreaQueue) Push(zone *api.ExploreResponse) {
	select {
	case q.ch <- zone:
		return
	default:
	}

	p := float32(zone.Amount) / float32(zone.Area.SizeX*zone.Area.SizeY)
	i := &Item{
		priority: int(p * 100000),
		value:    zone,
	}
	q.l.Lock()
	defer q.l.Unlock()
	heap.Push(q.pq, i)
}

func (q *AreaQueue) Peek() *api.ExploreResponse {
	q.l.Lock()
	if q.pq.Len() == 0 {
		q.l.Unlock()
		return <-q.ch
	}
	defer q.l.Unlock()
	result := heap.Pop(q.pq).(*Item).value
	return result
}
