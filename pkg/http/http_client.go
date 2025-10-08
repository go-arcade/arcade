package http

import (
	"github.com/go-resty/resty/v2"
	"io"
)

/**
 * @author: gagral.x@gmail.com
 * @time: 2024/11/6 20:39
 * @file: http_client.go
 * @description: http client
 */

type Request struct {
	Url     string
	Method  string
	Headers map[string]string
	Body    io.Reader
}

func NewRequest(url, method string, headers map[string]string, body io.Reader) *Request {
	return &Request{
		Url:     url,
		Method:  method,
		Headers: headers,
		Body:    body,
	}
}

func (r *Request) GET() *resty.Response {
	response, err := client().R().
		SetHeaders(r.Headers).
		Get(r.Url)
	if err != nil {
		return &resty.Response{}
	}
	return response
}

func (r *Request) POST() *resty.Response {
	response, err := client().R().
		SetHeaders(r.Headers).
		SetBody(r.Body).
		Post(r.Url)
	if err != nil {
		return &resty.Response{}
	}
	return response
}

func client() *resty.Client {
	return resty.New()
}
