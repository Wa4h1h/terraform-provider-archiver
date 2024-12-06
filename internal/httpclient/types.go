package httpclient

import (
	"context"
	"io"
	"net/http"
	"time"
)

type HTTPRunner interface {
	Do(ctx context.Context, params *RequestParams) (io.ReadCloser, error)
}

type HTTPHeaders map[string]string

type HTTPClientOpt func(*client)

type client struct {
	headers    HTTPHeaders
	timeout    *time.Duration
	hostname   string
	httpClient *http.Client
}

type RequestParams struct {
	Method  string
	Headers HTTPHeaders
	Body    io.ReadCloser
	Timeout int
}
