package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jamesnicolas/dirtcloud/domain"
)

// Client provides a Go SDK for the DirtCloud API
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	
	// Retry configuration
	retryMax              int
	retryInitialBackoffMs int
}

// Config holds client configuration
type Config struct {
	BaseURL               string
	Token                 string
	HTTPClient            *http.Client
	RetryMax              int
	RetryInitialBackoffMs int
}

// NewClient creates a new DirtCloud API client
func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8080"
	}
	
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	
	if config.RetryMax == 0 {
		config.RetryMax = 3
	}
	
	if config.RetryInitialBackoffMs == 0 {
		config.RetryInitialBackoffMs = 1000
	}
	
	return &Client{
		baseURL:               strings.TrimRight(config.BaseURL, "/"),
		token:                 config.Token,
		httpClient:            config.HTTPClient,
		retryMax:              config.RetryMax,
		retryInitialBackoffMs: config.RetryInitialBackoffMs,
	}
}

// do performs an HTTP request with retry logic
func (c *Client) do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	
	if body != nil {
		if s, ok := body.(string); ok {
			// Handle plain text body (for metadata)
			reqBody = strings.NewReader(s)
		} else {
			// Handle JSON body
			jsonData, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonData)
		}
	}
	
	url := c.baseURL + "/v1" + path
	
	var lastErr error
	backoff := time.Duration(c.retryInitialBackoffMs) * time.Millisecond
	
	for attempt := 0; attempt <= c.retryMax; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2 // Exponential backoff
		}
		
		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		
		// Set headers
		if body != nil {
			if _, ok := body.(string); ok {
				req.Header.Set("Content-Type", "text/plain")
			} else {
				req.Header.Set("Content-Type", "application/json")
			}
		}
		
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
		
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		
		defer resp.Body.Close()
		
		// Check if we should retry
		if shouldRetry(resp.StatusCode) {
			respBody, _ := io.ReadAll(resp.Body)
			
			// Handle Retry-After header for 429 responses
			if resp.StatusCode == http.StatusTooManyRequests {
				if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
					if seconds, err := strconv.Atoi(retryAfter); err == nil {
						backoff = time.Duration(seconds) * time.Second
					}
				}
			}
			
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
			continue
		}
		
		// Handle successful responses
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if result != nil {
				if s, ok := result.(*string); ok {
					// Handle plain text response (for metadata)
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return fmt.Errorf("failed to read response body: %w", err)
					}
					*s = string(body)
				} else {
					// Handle JSON response
					if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
						return fmt.Errorf("failed to decode response: %w", err)
					}
				}
			}
			return nil
		}
		
		// Handle client/server errors (don't retry)
		respBody, _ := io.ReadAll(resp.Body)
		
		// Try to decode as DirtError
		var dirtErr domain.DirtError
		if err := json.Unmarshal(respBody, &dirtErr); err == nil {
			return &dirtErr
		}
		
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	return lastErr
}

// shouldRetry determines if a request should be retried based on status code
func shouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

// Project operations

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, req domain.CreateProjectRequest) (*domain.Project, error) {
	var project domain.Project
	err := c.do(ctx, "POST", "/projects", req, &project)
	return &project, err
}

// GetProject retrieves a project by ID
func (c *Client) GetProject(ctx context.Context, id string) (*domain.Project, error) {
	var project domain.Project
	err := c.do(ctx, "GET", "/projects/"+url.PathEscape(id), nil, &project)
	return &project, err
}

// ListProjects lists projects with optional filtering
func (c *Client) ListProjects(ctx context.Context, opts domain.ProjectListOptions) ([]*domain.Project, error) {
	path := "/projects"
	if opts.Name != "" {
		path += "?name=" + url.QueryEscape(opts.Name)
	}
	
	var projects []*domain.Project
	err := c.do(ctx, "GET", path, nil, &projects)
	return projects, err
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(ctx context.Context, id string, req domain.UpdateProjectRequest) (*domain.Project, error) {
	var project domain.Project
	err := c.do(ctx, "PATCH", "/projects/"+url.PathEscape(id), req, &project)
	return &project, err
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", "/projects/"+url.PathEscape(id), nil, nil)
}

// Instance operations

// CreateInstance creates a new instance
func (c *Client) CreateInstance(ctx context.Context, req domain.CreateInstanceRequest) (*domain.Instance, error) {
	var instance domain.Instance
	err := c.do(ctx, "POST", "/instances", req, &instance)
	return &instance, err
}

// GetInstance retrieves an instance by ID
func (c *Client) GetInstance(ctx context.Context, id string) (*domain.Instance, error) {
	var instance domain.Instance
	err := c.do(ctx, "GET", "/instances/"+url.PathEscape(id), nil, &instance)
	return &instance, err
}

// ListInstances lists instances with optional filtering
func (c *Client) ListInstances(ctx context.Context, opts domain.InstanceListOptions) ([]*domain.Instance, error) {
	path := "/instances"
	params := url.Values{}
	
	if opts.ProjectID != "" {
		params.Set("project_id", opts.ProjectID)
	}
	if opts.Name != "" {
		params.Set("name", opts.Name)
	}
	if opts.Status != "" {
		params.Set("status", opts.Status)
	}
	
	if len(params) > 0 {
		path += "?" + params.Encode()
	}
	
	var instances []*domain.Instance
	err := c.do(ctx, "GET", path, nil, &instances)
	return instances, err
}

// UpdateInstance updates an existing instance
func (c *Client) UpdateInstance(ctx context.Context, id string, req domain.UpdateInstanceRequest) (*domain.Instance, error) {
	var instance domain.Instance
	err := c.do(ctx, "PATCH", "/instances/"+url.PathEscape(id), req, &instance)
	return &instance, err
}

// DeleteInstance deletes an instance
func (c *Client) DeleteInstance(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", "/instances/"+url.PathEscape(id), nil, nil)
}

// Metadata operations

// CreateMetadata creates new metadata
func (c *Client) CreateMetadata(ctx context.Context, req domain.CreateMetadataRequest) (*domain.Metadata, error) {
	var metadata domain.Metadata
	err := c.do(ctx, "POST", "/metadata", req, &metadata)
	return &metadata, err
}

// GetMetadata retrieves metadata by ID
func (c *Client) GetMetadata(ctx context.Context, id string) (*domain.Metadata, error) {
	var metadata domain.Metadata
	err := c.do(ctx, "GET", "/metadata/"+id, nil, &metadata)
	return &metadata, err
}

// UpdateMetadata updates existing metadata
func (c *Client) UpdateMetadata(ctx context.Context, id string, req domain.UpdateMetadataRequest) (*domain.Metadata, error) {
	var metadata domain.Metadata
	err := c.do(ctx, "PATCH", "/metadata/"+id, req, &metadata)
	return &metadata, err
}

// ListMetadata lists metadata with optional prefix filtering
func (c *Client) ListMetadata(ctx context.Context, opts domain.MetadataListOptions) ([]*domain.Metadata, error) {
	path := "/metadata"
	if opts.Prefix != "" {
		path += "?prefix=" + url.QueryEscape(opts.Prefix)
	}
	
	var metadata []*domain.Metadata
	err := c.do(ctx, "GET", path, nil, &metadata)
	return metadata, err
}

// DeleteMetadata deletes metadata by ID
func (c *Client) DeleteMetadata(ctx context.Context, id string) error {
	return c.do(ctx, "DELETE", "/metadata/"+id, nil, nil)
}