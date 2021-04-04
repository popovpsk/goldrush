package api

import (
	"fmt"
	"goldrush/metrics"
	"sync/atomic"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

func TestGateway_Do(t *testing.T) {
	gw := NewGateway(metrics.NewMetricsSvc())
	body := "ok"

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/health":
			ctx.Response.SetStatusCode(200)
			ctx.Response.SetBodyString(body)
		default:
			ctx.Error("not found", fasthttp.StatusNotFound)
		}
	}

	go func() {
		err := fasthttp.ListenAndServe(":12345", m)
		if err != nil {
			t.Fatal(err)
		}
	}()

	time.Sleep(time.Millisecond * 10)

	t.Log("test server up")

	var counter int32

	for i := 0; i < 50; i++ {
		go func() {
			for {
				req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
				req.SetRequestURI("http://localhost:12345/health")
				err := gw.Do(req, res, 0)
				if err != nil {
					fmt.Println(err)
					t.Fail()
				}
				if res.StatusCode() == 200 && string(res.Body()) == body {
					atomic.AddInt32(&counter, 1)
				} else {
					t.Fail()
				}
				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(res)
			}
		}()
	}
	time.Sleep(time.Second)

	atomic.StoreInt32(&counter, 0)
	time.Sleep(time.Second)
	res := atomic.LoadInt32(&counter)
	fmt.Println(res, rps)
}
