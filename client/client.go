package client

import (
	"bytes"
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

type HTTPClient interface {
	Do(ctx context.Context, method string, handler string, body []byte) (int, []byte, error)
	Get(ctx context.Context, handler string) (int, []byte, error)
	Post(ctx context.Context, handler string, body []byte) (int, []byte, error)
	Put(ctx context.Context, handler string, body []byte) (int, []byte, error)
	Delete(ctx context.Context, handler string, body []byte) (int, []byte, error)
}

type client struct {
	appName     string
	host        string
	http        *http.Client
	middlewares []Middleware
	closers     map[string]func() error
	logger      zerolog.Logger
	tracer      opentracing.Tracer
}

func New(appName, host string, options ...Option) HTTPClient {
	c := &client{
		http: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   defaultTimeout,
		},
		appName:     appName,
		host:        strings.TrimSuffix(host, "/"),
		middlewares: []Middleware{},
		closers:     map[string]func() error{},
		logger:      zerolog.Nop(),
	}

	for _, o := range options {
		o(c)
	}

	return c
}

func (c *client) Do(ctx context.Context, method string, handler string, body []byte) (int, []byte, error) {
	defer c.close()

	req, err := http.NewRequestWithContext(ctx, method, c.host+handler, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}

	// middlewares
	for _, mv := range c.middlewares {
		mv(ctx, req)
	}

	// tracing
	if c.tracer != nil {
		span, _ := opentracing.StartSpanFromContext(ctx, handler)
		defer span.Finish()
		err = c.tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		if err != nil {
			return 0, nil, err
		}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, respData, nil
}

func (c *client) Get(ctx context.Context, handler string) (int, []byte, error) {
	return c.Do(ctx, http.MethodGet, handler, nil)
}

func (c *client) Post(ctx context.Context, handler string, body []byte) (int, []byte, error) {
	return c.Do(ctx, http.MethodPost, handler, body)
}

func (c *client) Put(ctx context.Context, handler string, body []byte) (int, []byte, error) {
	return c.Do(ctx, http.MethodPut, handler, body)
}

func (c *client) Delete(ctx context.Context, handler string, body []byte) (int, []byte, error) {
	return c.Do(ctx, http.MethodDelete, handler, body)
}

func (c *client) close() {
	for name, closer := range c.closers {
		if err := closer(); err != nil {
			c.logger.Warn().Err(err).Msgf("can't close %s", name)
		}
	}
}
