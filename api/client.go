package api

import (
	"fmt"
	"goldrush/types"
	"sync/atomic"

	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

const contentTypeJson = "application/json"
const headerAccept = "*/*"
const headerAcceptEncoding = "gzip, deflate"

type Client struct {
	cl             *Gateway
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

func NewClient(url string, gw *Gateway) *Client {
	return &Client{
		cl:             gw,
		exploreURI:     []byte(url + "/explore"),
		licensesURI:    []byte(url + "/licenses"),
		digURI:         []byte(url + "/dig"),
		cashURI:        []byte(url + "/cash"),
		emptyArrayBody: []byte("[]"),
	}
}

func (c *Client) Explore(area *types.Area, response *types.ExploreResponse) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	c.AddHeaders(req)
	req.SetRequestURIBytes(c.exploreURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	easyjson.MarshalToWriter(area, req.BodyWriter())
	for {
		c.cl.Do(req, res, 0)
		if res.StatusCode() == 200 {
			break
		}
		atomic.AddInt32(&ErcntExp, 1)
	}
	easyjson.Unmarshal(res.Body(), response)
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func (c *Client) PostLicenses(wallet types.PostLicenseRequest, response *types.License) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.licensesURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	c.AddHeaders(req)

	if wallet != nil {
		easyjson.MarshalToWriter(wallet, req.BodyWriter())
	} else {
		req.SetBodyRaw(c.emptyArrayBody)
	}

	for {
		c.cl.Do(req, res, 3)
		if res.StatusCode() == 200 {
			break
		}
		atomic.AddInt32(&ErcntLic, 1)
	}
	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		fmt.Println(err)
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
	return
}

func (c *Client) Dig(request *types.DigRequest, response *types.Treasures) bool {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.digURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	c.AddHeaders(req)

	easyjson.MarshalToWriter(request, req.BodyWriter())

	for {
		c.cl.Do(req, res, 0)
		if res.StatusCode() == 404 {
			return false
		}
		if res.StatusCode() != 200 {
			fmt.Println(string(res.Body()))
		}
		if res.StatusCode() == 200 {
			break
		}
		atomic.AddInt32(&ErrcntDig, 1)
	}
	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		fmt.Println(err)
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
	return true
}

func (c *Client) Cash(request string, response *types.Payment) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.cashURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	c.AddHeaders(req)

	req.SetBodyRaw([]byte(fmt.Sprintf("\"%s\"", request)))
	for {
		c.cl.Do(req, res, 3)
		if res.StatusCode() == 200 {
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

func (c *Client) AddHeaders(req *fasthttp.Request) {
	req.Header.SetContentType(contentTypeJson)
	req.Header.Add(fasthttp.HeaderAccept, headerAccept)
	req.Header.Add(fasthttp.HeaderAcceptEncoding, headerAcceptEncoding)
}
