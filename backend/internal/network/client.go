package network

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 14.7; rv:135.0) Gecko/20100101 Firefox/135.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36 Edg/132.0.0.0",
}

// RandomUserAgent returns a modern browser User-Agent string.
func RandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// Client wraps an http.Client with connection pooling, proxy support, and helper methods.
type Client struct {
	httpClient *http.Client
	transport  *http.Transport
}

// ClientOptions contains configuration options for the network client.
type ClientOptions struct {
	Timeout         time.Duration
	MaxIdleConns    int
	MaxConnsPerHost int
	InsecureSkipTLS bool
	ProxyURL        string
}

// NewClient creates a new configured network Client.
func NewClient(opts ClientOptions) *Client {
	if opts.Timeout == 0 {
		opts.Timeout = 3 * time.Second
	}
	if opts.MaxIdleConns == 0 {
		opts.MaxIdleConns = 500
	}
	if opts.MaxConnsPerHost == 0 {
		opts.MaxConnsPerHost = 100
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   800 * time.Millisecond,
			KeepAlive: 90 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          opts.MaxIdleConns,
		MaxIdleConnsPerHost:   opts.MaxConnsPerHost,
		MaxConnsPerHost:       opts.MaxConnsPerHost,
		IdleConnTimeout:       120 * time.Second,
		TLSHandshakeTimeout:   800 * time.Millisecond,
		ResponseHeaderTimeout: 1500 * time.Millisecond,
		ExpectContinueTimeout: 500 * time.Millisecond,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: opts.InsecureSkipTLS,
		},
	}

	if opts.ProxyURL != "" {
		if strings.HasPrefix(opts.ProxyURL, "socks5://") {
			proxyAddr := strings.TrimPrefix(opts.ProxyURL, "socks5://")
			dialer, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
			if err == nil {
				transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				}
			}
		} else {
			if u, err := url.Parse(opts.ProxyURL); err == nil {
				transport.Proxy = http.ProxyURL(u)
			}
		}
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			return nil
		},
	}

	return &Client{
		httpClient: httpClient,
		transport:  transport,
	}
}

// RequestOptions defines parameters for outbound HTTP requests.
type RequestOptions struct {
	Headers     map[string]string
	Cookies     map[string]string
	Body        io.Reader
	ContentType string
	Timeout     time.Duration
}

// Get executes a GET request with standard headers and returns the response and body.
func (c *Client) Get(ctx context.Context, targetURL string, opts *RequestOptions) (*http.Response, []byte, error) {
	return c.Do(ctx, http.MethodGet, targetURL, opts)
}

// Post executes a POST request and returns the response and body.
func (c *Client) Post(ctx context.Context, targetURL string, opts *RequestOptions) (*http.Response, []byte, error) {
	return c.Do(ctx, http.MethodPost, targetURL, opts)
}

// Do executes a generic HTTP request.
func (c *Client) Do(ctx context.Context, method, targetURL string, opts *RequestOptions) (*http.Response, []byte, error) {
	var body io.Reader
	if opts != nil && opts.Body != nil {
		body = opts.Body
	}

	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default headers
	req.Header.Set("User-Agent", RandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8,application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("DNT", "1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")

	if opts != nil {
		if opts.ContentType != "" {
			req.Header.Set("Content-Type", opts.ContentType)
		}
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
		for k, v := range opts.Cookies {
			req.AddCookie(&http.Cookie{Name: k, Value: v})
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, fmt.Errorf("failed reading response body: %w", err)
	}

	return resp, respBody, nil
}
