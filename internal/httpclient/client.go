package httpclient

import (
	"context"
	"io"
	"net/http"
	"time"
)

// make sure we conform to HTTPRunner
var _ HTTPRunner = &client{}

func NewHTTPRunner(opts ...HTTPClientOpt) HTTPRunner {
	c := new(client)

	for _, opt := range opts {
		opt(c)
	}

	c.httpClient = new(http.Client)

	if c.timeout == nil {
		WithTimeout(DefaultTimeout)(c)
	}

	return c
}

func WithTimeout(timeout int) HTTPClientOpt {
	return func(c *client) {
		secDuration := time.Duration(timeout) * time.Second
		c.timeout = &secDuration
	}
}

func WithHeaders(headers HTTPHeaders) HTTPClientOpt {
	return func(c *client) {
		c.headers = headers
	}
}

func WithHostname(hostname string) HTTPClientOpt {
	return func(c *client) {
		c.hostname = hostname
	}
}

func (c *client) Do(ctx context.Context, params *RequestParams,
) (io.ReadCloser, error) {
	return nil, nil
}
