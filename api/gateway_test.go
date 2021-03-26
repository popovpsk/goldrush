package api

import (
	"github.com/valyala/fasthttp"
	"testing"
)

func TestGateway_Do(t *testing.T) {
	g := NewGateWay()
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURI("http://192.168.1.49:8090/")
	g.Do(req, res, 0)
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}
