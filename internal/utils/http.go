package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

// HTTPClient handles HTTP requests
type HTTPClient struct {
	baseURL string
	client  *resty.Client
	logger  *logrus.Logger
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string, logger *logrus.Logger) *HTTPClient {
	client := resty.New()
	client.SetBaseURL(baseURL)
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)
	client.SetRetryMaxWaitTime(5 * time.Second)

	return &HTTPClient{
		baseURL: baseURL,
		client:  client,
		logger:  logger,
	}
}

// MakeRequest makes an HTTP request
func (c *HTTPClient) MakeRequest(method, path string, params map[string]interface{}) (interface{}, error) {
	c.logger.WithFields(logrus.Fields{
		"method": method,
		"path":   path,
		"params": params,
	}).Debug("Making HTTP request")

	// Create request
	req := c.client.R()

	// Set headers
	req.SetHeader("Content-Type", "application/json")
	req.SetHeader("Accept", "application/json")

	// Handle different HTTP methods
	switch method {
	case "GET":
		return c.handleGET(req, path, params)
	case "POST":
		return c.handlePOST(req, path, params)
	case "PUT":
		return c.handlePUT(req, path, params)
	case "DELETE":
		return c.handleDELETE(req, path, params)
	case "PATCH":
		return c.handlePATCH(req, path, params)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}
}

// handleGET handles GET requests
func (c *HTTPClient) handleGET(req *resty.Request, path string, params map[string]interface{}) (interface{}, error) {
	// Add query parameters
	for key, value := range params {
		req.SetQueryParam(key, fmt.Sprintf("%v", value))
	}

	resp, err := req.Get(path)
	if err != nil {
		return nil, fmt.Errorf("GET request failed: %w", err)
	}

	return c.parseResponse(resp)
}

// handlePOST handles POST requests
func (c *HTTPClient) handlePOST(req *resty.Request, path string, params map[string]interface{}) (interface{}, error) {
	// Set request body
	if body, exists := params["body"]; exists {
		req.SetBody(body)
		delete(params, "body")
	}

	// Add remaining parameters as query parameters
	for key, value := range params {
		req.SetQueryParam(key, fmt.Sprintf("%v", value))
	}

	resp, err := req.Post(path)
	if err != nil {
		return nil, fmt.Errorf("POST request failed: %w", err)
	}

	return c.parseResponse(resp)
}

// handlePUT handles PUT requests
func (c *HTTPClient) handlePUT(req *resty.Request, path string, params map[string]interface{}) (interface{}, error) {
	// Set request body
	if body, exists := params["body"]; exists {
		req.SetBody(body)
		delete(params, "body")
	}

	// Add remaining parameters as query parameters
	for key, value := range params {
		req.SetQueryParam(key, fmt.Sprintf("%v", value))
	}

	resp, err := req.Put(path)
	if err != nil {
		return nil, fmt.Errorf("PUT request failed: %w", err)
	}

	return c.parseResponse(resp)
}

// handleDELETE handles DELETE requests
func (c *HTTPClient) handleDELETE(req *resty.Request, path string, params map[string]interface{}) (interface{}, error) {
	// Add query parameters
	for key, value := range params {
		req.SetQueryParam(key, fmt.Sprintf("%v", value))
	}

	resp, err := req.Delete(path)
	if err != nil {
		return nil, fmt.Errorf("DELETE request failed: %w", err)
	}

	return c.parseResponse(resp)
}

// handlePATCH handles PATCH requests
func (c *HTTPClient) handlePATCH(req *resty.Request, path string, params map[string]interface{}) (interface{}, error) {
	// Set request body
	if body, exists := params["body"]; exists {
		req.SetBody(body)
		delete(params, "body")
	}

	// Add remaining parameters as query parameters
	for key, value := range params {
		req.SetQueryParam(key, fmt.Sprintf("%v", value))
	}

	resp, err := req.Patch(path)
	if err != nil {
		return nil, fmt.Errorf("PATCH request failed: %w", err)
	}

	return c.parseResponse(resp)
}

// parseResponse parses the HTTP response
func (c *HTTPClient) parseResponse(resp *resty.Response) (interface{}, error) {
	c.logger.WithFields(logrus.Fields{
		"status_code": resp.StatusCode(),
		"size":        len(resp.Body()),
	}).Debug("Received HTTP response")

	// Check for HTTP errors
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode(), resp.String())
	}

	// Try to parse as JSON
	var result interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		// If JSON parsing fails, return the raw string
		return string(resp.Body()), nil
	}

	return result, nil
}

// SetAuth sets authentication for the client
func (c *HTTPClient) SetAuth(authType, token string) {
	switch authType {
	case "bearer":
		c.client.SetAuthToken(token)
	case "apikey":
		c.client.SetHeader("X-API-Key", token)
	case "basic":
		// TODO: Implement basic auth
		c.logger.Warn("Basic authentication not implemented")
	default:
		c.logger.Warnf("Unknown authentication type: %s", authType)
	}
}

// SetBaseURL sets the base URL for the client
func (c *HTTPClient) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
	c.client.SetBaseURL(baseURL)
}
