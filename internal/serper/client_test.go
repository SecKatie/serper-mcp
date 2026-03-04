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

package serper

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := &Client{
		apiKey:     "test-key",
		httpClient: srv.Client(),
	}
	return c, srv
}

// TestSearch_Success verifies that a 200 response is returned as-is.
func TestSearch_Success(t *testing.T) {
	want := `{"organic":[{"title":"Go","link":"https://go.dev"}]}`
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/search", r.URL.Path)
		require.Equal(t, "test-key", r.Header.Get("X-API-KEY"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(want))
	})

	// Override the hardcoded searchURL by sending directly via do.
	payload := []byte(`{"q":"golang"}`)
	got, err := c.do(context.Background(), srv.URL+"/search", payload)
	require.NoError(t, err)
	require.JSONEq(t, want, string(got))
}

// TestSearch_HTTPError verifies that non-200 responses produce a wrapped error.
func TestSearch_HTTPError(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})

	payload := []byte(`{"q":"golang"}`)
	_, err := c.do(context.Background(), srv.URL+"/search", payload)
	require.Error(t, err)
	require.Contains(t, err.Error(), "500")
}

// TestScrape_Success verifies that a 200 scrape response is returned as-is.
func TestScrape_Success(t *testing.T) {
	want := `{"text":"hello world"}`
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/scrape", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(want))
	})

	// Patch the scrapeURL constant via do directly.
	payload := []byte(`{"url":"https://example.com"}`)
	got, err := c.do(context.Background(), srv.URL+"/scrape", payload)
	require.NoError(t, err)
	require.JSONEq(t, want, string(got))
}

// TestScrape_InvalidURL verifies that non-http(s) and malformed URLs are rejected.
func TestScrape_InvalidURL(t *testing.T) {
	c := &Client{apiKey: "test-key", httpClient: http.DefaultClient}

	cases := []string{
		"ftp://example.com",
		"",
		"example.com",
		"//example.com",
	}

	for _, rawURL := range cases {
		t.Run(rawURL, func(t *testing.T) {
			_, err := c.Scrape(context.Background(), rawURL)
			require.Error(t, err)
			require.Contains(t, err.Error(), "invalid URL")
		})
	}
}

// TestScrape_BodySizeLimit verifies that responses larger than 10 MB are truncated without panic.
func TestScrape_BodySizeLimit(t *testing.T) {
	// Build a body larger than maxBodyBytes (10 MB).
	bigBody := bytes.Repeat([]byte("x"), maxBodyBytes+1024)

	c, srv := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(bigBody)
	})

	payload := []byte(`{"url":"https://example.com"}`)
	got, err := c.do(context.Background(), srv.URL, payload)

	// The response is technically not valid JSON when truncated, but the important
	// behaviour is: no panic, body length is capped at maxBodyBytes.
	require.NoError(t, err)
	require.LessOrEqual(t, len(got), maxBodyBytes, "response body must not exceed the cap")
	require.True(t, strings.HasPrefix(string(got), "x"), "body should start with the expected content")
}
