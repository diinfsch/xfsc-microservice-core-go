// Package cache implements a client of the Cache service.
package cache

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	errors "github.com/eclipse-xfsc/microservice-core-go/pkg/err"
)

// Client for the Cache service.
type Client struct {
	addr       string
	httpClient *http.Client
}

func New(addr string, opts ...Option) *Client {
	c := &Client{
		addr:       addr,
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Set(ctx context.Context, key, namespace, scope string, value []byte) error {
	requestURI := c.addr + "/v1/cache"
	cacheURL, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return errors.New(errors.Internal, "invalid cache url", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cacheURL.String(), bytes.NewReader(value))
	if err != nil {
		return err
	}

	req.Header = http.Header{
		"x-cache-key":       []string{key},
		"x-cache-namespace": []string{namespace},
		"x-cache-scope":     []string{scope},
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpected response: %s", resp.Status)
		return errors.New(errors.GetKind(resp.StatusCode), msg)
	}

	return nil
}

func (c *Client) Get(ctx context.Context, key, namespace, scope string) ([]byte, error) {
	requestURI := c.addr + "/v1/cache"
	cacheURL, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return nil, errors.New(errors.Internal, "invalid cache url", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", cacheURL.String(), nil)
	req.Header = http.Header{
		"x-cache-key":       []string{key},
		"x-cache-namespace": []string{namespace},
		"x-cache-scope":     []string{scope},
	}
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, errors.New(errors.NotFound)
		}
		msg := fmt.Sprintf("unexpected response: %s", resp.Status)
		return nil, errors.New(errors.GetKind(resp.StatusCode), msg)
	}

	return io.ReadAll(resp.Body)
}
