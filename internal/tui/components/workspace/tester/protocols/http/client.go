package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// TestResponse holds the result of the HTTP test request.
type TestResponse struct {
	StatusCode int
	StatusText string
	Body       string
	Headers    http.Header
	Latency    time.Duration
	IsBinary   bool
	Error      error
}

// Client executes HTTP tests safely.
type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	c := &Client{}
	c.applyDefaults(false, 10*time.Second)
	return c
}

// NewClientWithSettings creates a Client configured with the given settings.
func NewClientWithSettings(followRedirects bool, timeout time.Duration) *Client {
	c := &Client{}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	c.applyDefaults(followRedirects, timeout)
	return c
}

// applyDefaults sets up the underlying http.Client with the given settings.
func (c *Client) applyDefaults(followRedirects bool, timeout time.Duration) {
	checkRedirect := func(req *http.Request, via []*http.Request) error {
		if !followRedirects {
			return http.ErrUseLastResponse // Stop after first redirect
		}
		if len(via) >= 10 {
			return fmt.Errorf("stopped after 10 redirects")
		}
		return nil
	}
	c.httpClient = &http.Client{
		Timeout:       timeout,
		CheckRedirect: checkRedirect,
		Transport: &http.Transport{
			ForceAttemptHTTP2: false,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402 -- intentional for local mock testing
			},
		},
	}
}

// ApplySettings reconfigures the Client based on live SettingsConfig.
// Called whenever the user closes the Settings overlay or at startup.
func (c *Client) ApplySettings(followRedirects bool, timeout time.Duration) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	c.applyDefaults(followRedirects, timeout)
}

// Timeout returns the currently configured HTTP timeout.
func (c *Client) Timeout() time.Duration {
	if c.httpClient != nil {
		return c.httpClient.Timeout
	}
	return 0
}

// FollowRedirectsEnabled returns true if the client is configured to follow HTTP redirects.
func (c *Client) FollowRedirectsEnabled() bool {
	if c.httpClient == nil || c.httpClient.CheckRedirect == nil {
		return false
	}
	err := c.httpClient.CheckRedirect(nil, nil)
	return err == nil
}

// ExecuteRequest runs the request and prevents leaks/DoS.
func (c *Client) ExecuteRequest(ctx context.Context, method, url string, headers string, rawBody string) *TestResponse {
	start := time.Now()
	
	// Create request
	var bodyReader io.Reader
	if rawBody != "" {
		bodyReader = bytes.NewReader([]byte(rawBody))
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return &TestResponse{Error: err}
	}
	
	// Parse manual headers (Key: Value)
	if headers != "" {
		lines := strings.Split(headers, "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				req.Header.Add(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
			}
		}
	}
	
	// Auto-infer JSON if no content-type provided
	if req.Header.Get("Content-Type") == "" {
		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
			req.Header.Set("Content-Type", "application/json")
		} else if rawBody != "" {
			if strings.HasPrefix(strings.TrimSpace(rawBody), "{") || strings.HasPrefix(strings.TrimSpace(rawBody), "[") {
				req.Header.Set("Content-Type", "application/json")
			}
		}
	}

	// Execute
	resp, err := c.httpClient.Do(req)
	latency := time.Since(start)
	
	if err != nil {
		return &TestResponse{Error: err, Latency: latency}
	}
	defer resp.Body.Close()

	// PKS-008: Binary Protection (Max 1MB)
	contentType := resp.Header.Get("Content-Type")
	contentLength := resp.ContentLength
	
	isBinary := false
	if strings.HasPrefix(contentType, "image/") || strings.HasPrefix(contentType, "application/pdf") || contentLength > 1024*1024 {
		isBinary = true
	}

	var bodyStr string
	if isBinary {
		bodyStr = fmt.Sprintf("[ Binary Data - Content-Length: %d ]", contentLength)
	} else {
		// Limit reader to 1MB to prevent chunked DoS
		lr := io.LimitReader(resp.Body, 1024*1024)
		bodyBytes, err := io.ReadAll(lr)
		if err != nil && err != io.EOF {
			bodyStr = "Error reading body: " + err.Error()
		} else {
			bodyStr = string(bodyBytes)
		}
	}

	return &TestResponse{
		StatusCode: resp.StatusCode,
		StatusText: resp.Status,
		Headers:    resp.Header,
		Body:       bodyStr,
		Latency:    latency,
		IsBinary:   isBinary,
	}
}
