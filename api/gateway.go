package api

import (
	"github.com/valyala/fasthttp"
	"sync/atomic"

	"time"
)

type request struct {
	req  *fasthttp.Request
	res  *fasthttp.Response
	done chan struct{}
}

type Gateway struct {
	requests chan *request
	slaveCh  chan *request
	pool     chan *request
	cl       *fasthttp.Client
	t        time.Time

	next     *request
	heavyCnt int32
	parallel int32
	wakeUp   chan struct{}
}

func NewGateWay() *Gateway {
	g := &Gateway{
		requests: make(chan *request, 100),
		pool:     make(chan *request, 100),
		slaveCh:  make(chan *request, slaves),
		cl:       &fasthttp.Client{},
		t:        time.Now(),
		wakeUp:   make(chan struct{}),
	}
	for i := 0; i < 100; i++ {
		g.pool <- &request{
			done: make(chan struct{}),
		}
	}

	go g.startMasterWorker()
	for i := 0; i < slaves; i++ {
		go g.slaveWorker()
	}
	return g
}

const slaves = 10
const rpc = 1500
const parallelRequests = 3

const delay = 1000*1000/rpc*time.Microsecond - 10*time.Microsecond

func (g *Gateway) Do(req *fasthttp.Request, res *fasthttp.Response, t byte) {
	r := <-g.pool
	r.req = req
	r.res = res
	g.requests <- r
	<-r.done
}

func (g *Gateway) startMasterWorker() {
	for {
		r := <-g.requests
		g.waitDelay()
		if atomic.LoadInt32(&g.parallel) >= parallelRequests {
			<-g.wakeUp
		}
		atomic.AddInt32(&g.parallel, 1)
		g.slaveCh <- r
	}
}
func (g *Gateway) waitDelay() {
	ct := time.Now()
	if ct.Sub(g.t) > delay {
		g.t = ct
	} else {
		<-time.After(ct.Sub(g.t))
		g.t = time.Now()
	}
}

func (g *Gateway) slaveWorker() {
	for {
		r := <-g.slaveCh
		fasthttp.Do(r.req, r.res)
		atomic.AddInt32(&g.parallel, -1)
		select {
		case g.wakeUp <- struct{}{}:
		default:
		}

		r.done <- struct{}{}
		r.req = nil
		r.res = nil
		g.pool <- r
	}
}
