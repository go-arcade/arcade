package http

import (
	"context"
	"io"

	"github.com/go-arcade/arcade/pkg/trace"
	"github.com/go-arcade/arcade/pkg/trace/inject"
	"github.com/go-resty/resty/v2"
)

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
	return r.GETWithContext(context.Background())
}

func (r *Request) GETWithContext(ctx context.Context) *resty.Response {
	ctx = trace.ContextWithSpan(ctx)

	var response *resty.Response
	_, _, err := inject.HTTPRequest(ctx, r.Method, r.Url, func(ctx context.Context) (int, int64, error) {
		var reqErr error
		response, reqErr = client().R().
			SetContext(ctx).
			SetHeaders(r.Headers).
			Get(r.Url)
		if reqErr != nil {
			return 0, 0, reqErr
		}
		return response.StatusCode(), response.Size(), nil
	})

	if err != nil || response == nil {
		return &resty.Response{}
	}
	return response
}

func (r *Request) POST() *resty.Response {
	return r.POSTWithContext(context.Background())
}

func (r *Request) POSTWithContext(ctx context.Context) *resty.Response {
	ctx = trace.ContextWithSpan(ctx)

	var response *resty.Response
	_, _, err := inject.HTTPRequest(ctx, r.Method, r.Url, func(ctx context.Context) (int, int64, error) {
		var reqErr error
		response, reqErr = client().R().
			SetContext(ctx).
			SetHeaders(r.Headers).
			SetBody(r.Body).
			Post(r.Url)
		if reqErr != nil {
			return 0, 0, reqErr
		}
		return response.StatusCode(), response.Size(), nil
	})

	if err != nil || response == nil {
		return &resty.Response{}
	}
	return response
}

func client() *resty.Client {
	return resty.New()
}
