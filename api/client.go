package api

import (
	"fmt"
	"goldrush/metrics"
	"goldrush/types"
	"sync/atomic"

	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

const contentTypeJson = "application/json"
const headerAccept = "*/*"
const headerAcceptEncoding = "gzip, deflate"

type Client struct {
	m  *metrics.Svc
	cl *Gateway
	js *JsonS

	exploreURI     []byte
	licensesURI    []byte
	digURI         []byte
	cashURI        []byte
	emptyArrayBody []byte
}

var ErcntExp int32
var ErcntLic int32
var ErrcntDig int32
var ErrcntCash int32

func EndLog() {
	fmt.Printf("errors: exp:%v, lic:%v, dig:%v, cash:%v\n", ErcntExp, ErcntLic, ErrcntDig, ErrcntCash)
}

func NewClient(url string, gw *Gateway, m *metrics.Svc) *Client {

	return &Client{
		cl:             gw,
		m:              m,
		js:             NewJsonS(),
		exploreURI:     []byte(url + "/explore"),
		licensesURI:    []byte(url + "/licenses"),
		digURI:         []byte(url + "/dig"),
		cashURI:        []byte(url + "/cash"),
		emptyArrayBody: []byte("[]"),
	}
}

func (c *Client) Explore(area *types.Area, response *types.ExploredArea) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	c.addHeaders(req)
	req.SetRequestURIBytes(c.exploreURI)
	req.Header.SetMethod(fasthttp.MethodPost)

	body := c.js.GetExploreRequest(area)
	req.SetBodyRaw(*body)
	for {
		var err error

		//if area.Size() == 1 {
		//	err = c.cl.doTimeout(req, res, 1)
		//} else {
		//
		//}
		err = c.cl.Do(req, res, 0)
		if err == nil && res.StatusCode() == 200 {
			break
		}
		atomic.AddInt32(&ErcntExp, 1)
	}

	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		fmt.Println(err)
	}
	c.js.ReleaseExp(body)
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func (c *Client) PostLicenses(wallet types.PostLicenseRequest, response *types.License) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.licensesURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	c.addHeaders(req)
	if wallet != nil {
		easyjson.MarshalToWriter(wallet, req.BodyWriter())
	} else {
		req.SetBodyRaw(c.emptyArrayBody)
	}

	for {
		err := c.cl.Do(req, res, 21)
		if err == nil && res.StatusCode() == 200 {
			break
		}
		atomic.AddInt32(&ErcntLic, 1)
	}
	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		fmt.Println(err)
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func (c *Client) Dig(request *types.DigRequest, response *types.Treasures) bool {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.digURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	c.addHeaders(req)

	b := c.js.GetDigRequest(request)
	req.SetBodyRaw(*b)

	defer func() {
		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(res)
		c.js.ReleaseDig(b)
	}()

	for {
		err := c.cl.Do(req, res, 22)

		if res.StatusCode() == 404 {
			return false
		}
		if err == nil && res.StatusCode() == 200 {
			break
		}
		atomic.AddInt32(&ErrcntDig, 1)
	}
	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		fmt.Println(err)
	}
	return true
}

func (c *Client) Cash(request string, response *types.Payment) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.cashURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	c.addHeaders(req)

	req.SetBodyRaw([]byte(fmt.Sprintf("\"%s\"", request)))
	for {
		err := c.cl.Do(req, res, 23)
		if err == nil && res.StatusCode() == 200 {
			break
		}
		atomic.AddInt32(&ErrcntCash, 1)
	}
	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		fmt.Println(err.Error())
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func (c *Client) addHeaders(req *fasthttp.Request) {
	req.Header.SetContentType(contentTypeJson)
	req.Header.Add(fasthttp.HeaderAccept, headerAccept)
	req.Header.Add(fasthttp.HeaderAcceptEncoding, headerAcceptEncoding)
}
