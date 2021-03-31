package game

import (
	"sync"
)

type Worker struct {
	start func(state *int)
}

const (
	Working = 0
	Stopped = -1
	Slow    = 1
)

type Foreman struct {
	workers []func(state *int)
	pool    [][]*int
	l       sync.Mutex
}

func New() *Foreman {
	f := Foreman{
		workers: make([]func(state *int), 10),
		pool:    make([][]*int, 10),
	}
	for i := range f.pool {
		f.pool[i] = make([]*int, 0, 5)
	}
	return &f
}

func (f *Foreman) AddWorker(ID int, start func(state *int)) {
	f.l.Lock()
	defer f.l.Unlock()
	f.workers[ID] = start
}

func (f *Foreman) Start(ID, count int) {
	for i := 0; i < count; i++ {
		state := Working
		f.l.Lock()
		f.pool[ID] = append(f.pool[ID], &state)
		f.l.Unlock()

		go func() {
			fn := f.workers[ID]
			for {
				if state == Stopped {
					return
				}
				fn(&state)
			}
		}()
	}
}

func (f *Foreman) Stop(ID, count int) {
	f.l.Lock()
	defer f.l.Unlock()
	p := f.pool[ID]
	i := 0
	for _, v := range p {
		if *v != Stopped {
			*v = Stopped
			i++
			if i == count {
				return
			}
		}
	}
}

func (f *Foreman) StopAll(ID int) {
	f.l.Lock()
	defer f.l.Unlock()
	for _, v := range f.pool[ID] {
		*v = Stopped
	}
}

func (f *Foreman) ChangeState(ID, state, count int) {
	f.l.Lock()
	defer f.l.Unlock()
	p := f.pool[ID]
	i := 0
	for _, v := range p {
		if *v != Stopped {
			*v = state
			i++
			if i == count {
				return
			}
		}
	}
}
