package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/oauth2"
)

// Client is an HTTP client for the LibreGraph API with optional
// Host header override and DNS resolution override.
type Client struct {
	baseURL      string
	insecure     bool
	tokenSource  oauth2.TokenSource
	HostOverride string // HTTP Host header override
	ResolveIP    string // DNS resolution override (connect to this IP)
}

type Response struct {
	StatusCode int
	Body       string
}

func NewClient(baseURL string, insecure bool, ts oauth2.TokenSource) *Client {
	return &Client{baseURL: baseURL, insecure: insecure, tokenSource: ts}
}

// --- Transport ---

func (c *Client) newTransport() *http.Transport {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.insecure},
	}
	if c.ResolveIP != "" {
		d := &net.Dialer{}
		tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			_, port, _ := net.SplitHostPort(addr)
			if port == "" {
				port = "443"
			}
			return d.DialContext(ctx, network, net.JoinHostPort(c.ResolveIP, port))
		}
	}
	return tr
}

func (c *Client) applyHost(req *http.Request) {
	if c.HostOverride != "" {
		req.Host = c.HostOverride
	}
}

// --- Core request helpers ---

// request builds and executes an HTTP request with the client's
// transport, auth, and Host override. Returns the raw response.
func (c *Client) request(method, rawURL string, body io.Reader, contentLength int64, contentType string, extraHeaders map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, rawURL, body)
	if err != nil {
		return nil, fmt.Errorf("request create %s %s: %w", method, rawURL, err)
	}
	c.applyHost(req)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if contentLength > 0 {
		req.ContentLength = contentLength
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	slog.Debug("request", "method", method, "url", rawURL)
	return c.newClient().Do(req)
}

// newClient creates an http.Client with the configured transport.
func (c *Client) newClient() *http.Client {
	return &http.Client{Transport: c.newTransport()}
}

// getToken returns a Bearer token string from the token source.
func (c *Client) getToken() (string, error) {
	if c.tokenSource == nil {
		return "", fmt.Errorf("no token source configured")
	}
	t, err := c.tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("token: %w", err)
	}
	return t.AccessToken, nil
}

// --- Public API ---

// MakeRequest calls the LibreGraph API at path with the given method,
// JSON body, and query parameters. Returns a parsed Response.
func (c *Client) MakeRequest(path string, method string, body string, params url.Values) (*Response, error) {
	fullURL, err := url.JoinPath(c.baseURL, "graph", path)
	if err != nil {
		return nil, fmt.Errorf("MakeRequest join path %q: %w", path, err)
	}
	u, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("MakeRequest parse URL %q: %w", fullURL, err)
	}
	u.RawQuery = params.Encode()

	var r io.Reader
	headers := map[string]string{}
	if body != "" {
		r = bytes.NewBufferString(body)
		headers["Content-Type"] = "application/json"
	}

	token, err := c.getToken()
	if err != nil {
		return nil, fmt.Errorf("MakeRequest %s %s: %w", method, path, err)
	}
	headers["Authorization"] = "Bearer " + token

	resp, err := c.request(method, u.String(), r, int64(len(body)), "", headers)
	if err != nil {
		return nil, fmt.Errorf("MakeRequest %s %s: %w", method, path, err)
	}

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("MakeRequest %s %s: HTTP %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return readBody(resp)
}

// DoRequest is a low-level request helper used by the upload client.
// It returns the raw *http.Response for the caller to handle.
func (c *Client) DoRequest(method, rawURL string, body io.Reader, contentLength int64, contentType string, extraHeaders map[string]string) (*http.Response, error) {
	resp, err := c.request(method, rawURL, body, contentLength, contentType, extraHeaders)
	if err != nil {
		return nil, fmt.Errorf("DoRequest %s %s: %w", method, rawURL, err)
	}
	return resp, nil
}

// resolveURL resolves a potentially relative upload URL against baseURL.
func (c *Client) resolveURL(u string) string {
	if strings.HasPrefix(u, "http") {
		return u
	}
	return strings.TrimRight(c.baseURL, "/") + "/" + strings.TrimLeft(u, "/")
}

func readBody(resp *http.Response) (*Response, error) {
	defer resp.Body.Close()
	r := &Response{StatusCode: resp.StatusCode}
	if resp.Body == nil {
		return r, nil
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return r, fmt.Errorf("readBody: %w", err)
	}
	if len(b) > 0 {
		r.Body = string(b)
	}
	return r, nil
}
