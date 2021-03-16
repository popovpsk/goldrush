package goldpot

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type GoldPot struct {
	storage []string
	l       sync.Mutex
	ch      chan string
}

var (
	peekCntr  int32
	storeCntr int32
)

func EndLog() {
	fmt.Printf("GOLD: Store:%v, Peek:%v\n", atomic.LoadInt32(&storeCntr), atomic.LoadInt32(&peekCntr))
}

func New() *GoldPot {
	gp := &GoldPot{
		ch: make(chan string, 15000),
	}
	return gp
}

func (p *GoldPot) Store(gold []string) {
	atomic.AddInt32(&storeCntr, int32(len(gold)))
	for _, g := range gold {
		select {
		case p.ch <- g:
		default:
			go func() {
				p.ch <- g
			}()
		}
	}
}

func (p *GoldPot) Get() string {
	atomic.AddInt32(&peekCntr, 1)
	return <-p.ch
}
