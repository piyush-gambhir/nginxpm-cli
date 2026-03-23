package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/piyush-gambhir/nginxpm-cli/internal/build"
	"github.com/piyush-gambhir/nginxpm-cli/internal/config"
)

// Client is the Nginx Proxy Manager HTTP API client.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
	UserAgent  string
}

// NewClient creates a new client from a ResolvedConfig.
// It authenticates with the NPM API to obtain a JWT token.
func NewClient(rc *config.ResolvedConfig) (*Client, error) {
	if rc.URL == "" {
		return nil, fmt.Errorf("Nginx Proxy Manager URL is required (use --url, NGINXPM_URL, or configure a profile)")
	}

	baseURL := strings.TrimRight(rc.URL, "/")

	tlsConfig := buildTLSConfig(rc.Insecure)

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
		TLSClientConfig:     tlsConfig,
	}

	c := &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		UserAgent: "nginxpm-cli/" + build.Version,
	}

	// Authenticate to get a JWT token.
	if rc.Email != "" && rc.Password != "" {
		token, err := c.authenticate(context.Background(), rc.Email, rc.Password)
		if err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		c.Token = token
	}

	return c, nil
}

// NewClientWithToken creates a client with a pre-existing JWT token (for login testing).
func NewClientWithToken(baseURL string, insecure bool) *Client {
	tlsConfig := buildTLSConfig(insecure)

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableKeepAlives:   true,
		TLSClientConfig:     tlsConfig,
	}

	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		UserAgent: "nginxpm-cli/" + build.Version,
	}
}

// authenticate obtains a JWT token from the NPM API.
func (c *Client) authenticate(ctx context.Context, email, password string) (string, error) {
	body := map[string]string{
		"identity": email,
		"secret":   password,
	}

	resp, err := c.doNoAuth(ctx, http.MethodPost, "/api/tokens", body)
	if err != nil {
		return "", err
	}

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := resp.JSON(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.Token == "" {
		return "", fmt.Errorf("empty token in response (2FA may be required)")
	}

	return tokenResp.Token, nil
}

// Get sends a GET request.
func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

// Post sends a POST request with a JSON body.
func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.do(ctx, http.MethodPost, path, body)
}

// Put sends a PUT request with a JSON body.
func (c *Client) Put(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.do(ctx, http.MethodPut, path, body)
}

// Delete sends a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil)
}

// PostMultipart sends a POST request with a raw body and custom content type.
func (c *Client) PostMultipart(ctx context.Context, path string, body io.Reader, contentType string) (*Response, error) {
	url := c.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Content-Type", contentType)
	c.setAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	return &Response{HTTPResponse: resp, RequestURL: url}, nil
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) (*Response, error) {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.setAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	return &Response{HTTPResponse: resp, RequestURL: url}, nil
}

// doNoAuth sends a request without the Authorization header (for token retrieval).
func (c *Client) doNoAuth(ctx context.Context, method, path string, body interface{}) (*Response, error) {
	url := c.BaseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	return &Response{HTTPResponse: resp, RequestURL: url}, nil
}

// setAuth sets the Bearer token authorization header on the request.
func (c *Client) setAuth(req *http.Request) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
}
