package api

import (
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	rps              = 2000
	parallelRequests = 5
	delay            = 1000 * 1000 / rps * time.Microsecond
)

type Gateway struct {
	client *fasthttp.Client
	sema   chan struct{}

	l         sync.Mutex
	timeStamp time.Time
}

func NewGateway() *Gateway {
	return &Gateway{
		sema:   make(chan struct{}, parallelRequests),
		client: &fasthttp.Client{},
	}
}

func (g *Gateway) Do(req *fasthttp.Request, res *fasthttp.Response) error {
	g.sema <- struct{}{}
	g.waitNext()
	err := g.client.Do(req, res)
	<-g.sema
	return err
}

func (g *Gateway) waitNext() {
	g.l.Lock()
	t := time.Now()

	if t.Sub(g.timeStamp) >= 0 {
		g.timeStamp = time.Now().Add(delay)
		g.l.Unlock()
	} else {
		d := g.timeStamp.Sub(t)
		g.timeStamp = g.timeStamp.Add(delay)
		g.l.Unlock()
		time.Sleep(d)
	}
}
