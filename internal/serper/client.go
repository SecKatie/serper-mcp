/*
Copyright © 2026 Katie Mulliken <katie@mulliken.net>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Package serper provides an HTTP client for the Serper API (search and scrape).
package serper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	searchURL    = "https://google.serper.dev/search"
	scrapeURL    = "https://scrape.serper.dev"
	maxBodyBytes = 10 << 20 // 10 MB
)

// Client is an HTTP client for the Serper API.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Serper API client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// do marshals payload, POSTs to endpoint, and returns the raw JSON response body.
func (c *Client) do(ctx context.Context, endpoint string, payload []byte) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-API-KEY", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req) //nolint:gosec // endpoints are hardcoded constants or validated before call
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("serper API returned %s: %s", resp.Status, body)
	}

	return json.RawMessage(body), nil
}

// Search performs a Google search via the Serper API and returns the raw JSON response.
func (c *Client) Search(ctx context.Context, query string) (json.RawMessage, error) {
	payload, err := json.Marshal(map[string]string{"q": query})
	if err != nil {
		return nil, fmt.Errorf("marshal search payload: %w", err)
	}

	result, err := c.do(ctx, searchURL, payload)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	return result, nil
}

// Scrape fetches and returns the content of a URL via the Serper scrape API.
func (c *Client) Scrape(ctx context.Context, rawURL string) (json.RawMessage, error) {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("invalid URL %q: must be http or https", rawURL)
	}

	payload, err := json.Marshal(map[string]string{"url": rawURL})
	if err != nil {
		return nil, fmt.Errorf("marshal scrape payload: %w", err)
	}

	result, err := c.do(ctx, scrapeURL, payload)
	if err != nil {
		return nil, fmt.Errorf("scrape: %w", err)
	}

	return result, nil
}
