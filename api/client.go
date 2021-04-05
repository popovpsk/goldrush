package api

import (
	"fmt"
	"goldrush/types"

	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

const contentTypeJson = "application/json"
const headerAccept = "*/*"
const headerAcceptEncoding = "gzip, deflate"

type Client struct {
	cl *Gateway
	js *JsonS

	exploreURI     []byte
	licensesURI    []byte
	digURI         []byte
	cashURI        []byte
	emptyArrayBody []byte
}

func NewClient(url string, gw *Gateway) *Client {
	return &Client{
		cl:             gw,
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
		err := c.cl.Do(req, res)
		if err == nil && res.StatusCode() == 200 {
			err = easyjson.Unmarshal(res.Body(), response)
			if err == nil {
				break
			}
		}
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
		err := c.cl.Do(req, res)
		if err == nil && res.StatusCode() == 200 {
			err = easyjson.Unmarshal(res.Body(), response)
			if err == nil {
				break
			}
		}
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
		err := c.cl.Do(req, res)
		if res.StatusCode() == 404 {
			return false
		}
		if err == nil && res.StatusCode() == 200 {
			err = easyjson.Unmarshal(res.Body(), response)
			if err == nil {
				break
			}
		}
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
		err := c.cl.Do(req, res)
		if err == nil && res.StatusCode() == 200 {
			err = easyjson.Unmarshal(res.Body(), response)
			if err == nil {
				break
			}
		}
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func (c *Client) addHeaders(req *fasthttp.Request) {
	req.Header.SetContentType(contentTypeJson)
	req.Header.Add(fasthttp.HeaderAccept, headerAccept)
	req.Header.Add(fasthttp.HeaderAcceptEncoding, headerAcceptEncoding)
}
