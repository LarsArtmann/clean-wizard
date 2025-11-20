package adapters

import (
	"context"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

// HTTPClient provides a modern HTTP client with built-in features
type HTTPClient struct {
	client *resty.Client
}

// NewHTTPClient creates a new HTTP client with sensible defaults
func NewHTTPClient() *HTTPClient {
	client := resty.New()

	// Set sensible defaults
	client.
		SetTimeout(30*time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1*time.Second).
		SetRetryMaxWaitTime(10*time.Second).
		SetHeader("User-Agent", "clean-wizard/1.0.0")

	return &HTTPClient{
		client: client,
	}
}

// WithTimeout sets the request timeout
func (hc *HTTPClient) WithTimeout(timeout time.Duration) *HTTPClient {
	hc.client.SetTimeout(timeout)
	return hc
}

// WithRetry configures retry behavior
func (hc *HTTPClient) WithRetry(count int, waitTime, maxWaitTime time.Duration) *HTTPClient {
	hc.client.SetRetryCount(count).
		SetRetryWaitTime(waitTime).
		SetRetryMaxWaitTime(maxWaitTime)
	return hc
}

// WithAuth sets authentication header
func (hc *HTTPClient) WithAuth(authType, token string) *HTTPClient {
	// Trim whitespace from inputs
	authType = strings.TrimSpace(authType)
	token = strings.TrimSpace(token)

	// Build authorization header value
	var value string
	if strings.EqualFold(authType, "Bearer") {
		value = "Bearer " + token
	} else if authType != "" {
		value = authType + " " + token
	} else {
		value = token
	}

	// Set header directly instead of using SetAuthToken to avoid double Bearer
	hc.client.SetHeader("Authorization", value)
	return hc
}

// WithHeader adds a default header
func (hc *HTTPClient) WithHeader(key, value string) *HTTPClient {
	hc.client.SetHeader(key, value)
	return hc
}

// Get performs HTTP GET request
func (hc *HTTPClient) Get(ctx context.Context, url string) (*HTTPResponse, error) {
	resp, err := hc.client.R().SetContext(ctx).Get(url)
	if err != nil {
		return nil, err
	}
	bodyBytes := resp.Body()
	return &HTTPResponse{
		StatusCode: resp.StatusCode(),
		RawBody:    bodyBytes,
		Body:       string(bodyBytes), // Keep for convenience
		Headers:    resp.Header(),
		Request:    resp.Request,
	}, nil
}

// Post performs HTTP POST request
func (hc *HTTPClient) Post(ctx context.Context, url string, body any) (*HTTPResponse, error) {
	resp, err := hc.client.R().SetBody(body).SetContext(ctx).Post(url)
	if err != nil {
		return nil, err
	}
	bodyBytes := resp.Body()
	return &HTTPResponse{
		StatusCode: resp.StatusCode(),
		RawBody:    bodyBytes,
		Body:       string(bodyBytes), // Keep for convenience
		Headers:    resp.Header(),
		Request:    resp.Request,
	}, nil
}

// Put performs HTTP PUT request
func (hc *HTTPClient) Put(ctx context.Context, url string, body any) (*HTTPResponse, error) {
	resp, err := hc.client.R().SetBody(body).SetContext(ctx).Put(url)
	if err != nil {
		return nil, err
	}
	bodyBytes := resp.Body()
	return &HTTPResponse{
		StatusCode: resp.StatusCode(),
		RawBody:    bodyBytes,
		Body:       string(bodyBytes), // Keep for convenience
		Headers:    resp.Header(),
		Request:    resp.Request,
	}, nil
}

// Delete performs HTTP DELETE request
func (hc *HTTPClient) Delete(ctx context.Context, url string) (*HTTPResponse, error) {
	resp, err := hc.client.R().SetContext(ctx).Delete(url)
	if err != nil {
		return nil, err
	}
	bodyBytes := resp.Body()
	return &HTTPResponse{
		StatusCode: resp.StatusCode(),
		RawBody:    bodyBytes,
		Body:       string(bodyBytes), // Keep for convenience
		Headers:    resp.Header(),
		Request:    resp.Request,
	}, nil
}

// HTTPResponse wraps resty response
type HTTPResponse struct {
	StatusCode int                 `json:"status_code"`
	RawBody    []byte              `json:"raw_body"`
	Body       string              `json:"body,omitempty"` // Keep for convenience, marked as optional
	Headers    map[string][]string `json:"headers"`
	Request    *resty.Request      `json:"request"`
}

// IsSuccess returns true if status code indicates success (2xx)
func (hr *HTTPResponse) IsSuccess() bool {
	return hr.StatusCode >= 200 && hr.StatusCode < 300
}

// IsError returns true if status code indicates error (4xx, 5xx)
func (hr *HTTPResponse) IsError() bool {
	return hr.StatusCode >= 400
}

// IsClientError returns true if status code indicates client error (4xx)
func (hr *HTTPResponse) IsClientError() bool {
	return hr.StatusCode >= 400 && hr.StatusCode < 500
}

// IsServerError returns true if status code indicates server error (5xx)
func (hr *HTTPResponse) IsServerError() bool {
	return hr.StatusCode >= 500
}
