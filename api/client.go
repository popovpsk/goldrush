package api

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/mailru/easyjson"
	"github.com/valyala/fasthttp"
)

const contentTypeJson = "application/json"

type Client struct {
	cl             *fasthttp.Client
	logger         *logrus.Logger
	exploreURI     []byte
	getBalanceURI  []byte
	licensesURI    []byte
	digURI         []byte
	cashURI        []byte
	emptyArrayBody []byte
}

func NewClient(url string, logger *logrus.Logger) *Client {
	return &Client{
		cl: &fasthttp.Client{
			ReadTimeout:  time.Second,
			WriteTimeout: time.Second,
		},
		logger:         logger,
		exploreURI:     []byte(url + "/explore"),
		getBalanceURI:  []byte(url + "/balance"),
		licensesURI:    []byte(url + "/licenses"),
		digURI:         []byte(url + "/dig"),
		cashURI:        []byte(url + "/cash"),
		emptyArrayBody: []byte("[]"),
	}
}

func (c *Client) Explore(area *Area, response *ExploreResponse) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.exploreURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType(contentTypeJson)
	easyjson.MarshalToWriter(area, req.BodyWriter())
	for {
		err := c.cl.Do(req, res)
		if err == nil && res.StatusCode() == 200 {
			break
		}
	}
	easyjson.Unmarshal(res.Body(), response)
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func (c *Client) PostLicenses(wallet *PostLicenseRequest, response *License) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.licensesURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType(contentTypeJson)
	req.Header.Add(fasthttp.HeaderAccept, "*/*")
	req.Header.Add(fasthttp.HeaderAcceptEncoding, "gzip, deflate")

	if wallet != nil {
		easyjson.MarshalToWriter(wallet, req.BodyWriter())
	} else {
		req.SetBodyRaw(c.emptyArrayBody)
	}

	for {
		err := c.cl.Do(req, res)
		if err == nil && res.StatusCode() == 200 {
			break
		}
	}
	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		c.logger.Error(err)
	}

	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
	return
}

func (c *Client) Dig(request *DigRequest, response *Treasures) bool {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.digURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType(contentTypeJson)
	req.Header.Add(fasthttp.HeaderAccept, "*/*")
	req.Header.Add(fasthttp.HeaderAcceptEncoding, "gzip, deflate")

	easyjson.MarshalToWriter(request, req.BodyWriter())

	for {
		err2 := c.cl.Do(req, res)
		if res.StatusCode() == 404 {
			return false
		}
		if res.StatusCode() != 200 {
			c.logger.Infoln(string(res.Body()))
		}
		if err2 == nil && res.StatusCode() == 200 {
			break
		}
	}
	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		c.logger.Error(err)
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
	return true
}

func (c *Client) Cash(request string, response *Payment) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.cashURI)
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType(contentTypeJson)
	req.Header.Add(fasthttp.HeaderAccept, "*/*")
	req.Header.Add(fasthttp.HeaderAcceptEncoding, "gzip, deflate")
	req.SetBodyRaw([]byte(fmt.Sprintf("\"%s\"", request)))
	for {
		err := c.cl.Do(req, res)
		if err == nil && res.StatusCode() == 200 {
			break
		}
	}
	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		fmt.Println(err.Error())
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func (c *Client) GetBalance(response *BalanceResponse) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.getBalanceURI)
	req.Header.SetContentType(contentTypeJson)

	c.cl.Do(req, res)

	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		c.logger.Error(err)
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}

func (c *Client) GetLicenses(response *LicensesResponse) {
	req, res := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.SetRequestURIBytes(c.licensesURI)
	fmt.Println("call get lic")
	for {
		err := c.cl.Do(req, res)
		if err == nil && res.StatusCode() == 200 {
			fmt.Println("call licenses res", string(res.Body()))
			break
		}
	}
	if err := easyjson.Unmarshal(res.Body(), response); err != nil {
		c.logger.Error(err)
	}
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(res)
}
