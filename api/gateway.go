package api

import (
	"goldrush/metrics"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	rps              = 2000
	parallelRequests = 3
	delay            = 1000 * 1000 / rps * time.Microsecond
)

type Gateway struct {
	m      *metrics.Svc
	client *fasthttp.Client
	sema   chan struct{}

	l         sync.Mutex
	timeStamp time.Time
}

func NewGateway(m *metrics.Svc) *Gateway {
	go func() {
		for {
			atomic.StoreInt64(&rpscntr, 0)
			time.Sleep(time.Second)
			r := atomic.LoadInt64(&rpscntr)
			m.AddInt("rps", r)
		}
	}()

	return &Gateway{
		m:      m,
		sema:   make(chan struct{}, parallelRequests),
		client: &fasthttp.Client{},
	}
}

func (g *Gateway) Do(req *fasthttp.Request, res *fasthttp.Response, b int32) error {
	var err error
	g.sema <- struct{}{}
	g.waitNext()
	t := time.Now()
	err = g.client.Do(req, res)
	el := time.Since(t)
	<-g.sema
	g.m.Add(getMethodName(b), el)
	atomic.AddInt64(&rpscntr, 1)
	return err
}

var rpscntr int64

func (g *Gateway) doTimeout(req *fasthttp.Request, res *fasthttp.Response, b int32) error {
	for {
		reqCp := fasthttp.AcquireRequest()
		resCp := fasthttp.AcquireResponse()
		req.CopyTo(reqCp)
		g.sema <- struct{}{}
		g.waitNext()
		t := time.Now()
		err := g.client.DoTimeout(reqCp, resCp, time.Millisecond*3)
		el := time.Since(t)
		<-g.sema
		if err == nil {
			g.m.Add(getMethodName(1), el)
			resCp.CopyTo(res)
		}
		fasthttp.ReleaseRequest(reqCp)
		fasthttp.ReleaseResponse(resCp)
		if err != fasthttp.ErrTimeout {
			return err
		}
	}
}

func (g *Gateway) waitNext() {
	//горутина занимает квант в который полетит запрос
	//если разница между нау и меткой больше нуля можно
	//исполняться и cдвигать  иначе cдвигаем и ждем

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

func getMethodName(id int32) string {
	switch id {
	case 1:
		return "explore 1"
	case 21:
		return "license"
	case 22:
		return "dig"
	case 23:
		return "cash"
	default:
		return ""
	}
}
