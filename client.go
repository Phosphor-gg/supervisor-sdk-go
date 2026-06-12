// Package supervisor provides an official Go client for the Supervisor content moderation API.
package supervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.supervisor.gg"

// Client is the Supervisor API client for content moderation.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) { c.httpClient = client }
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// NewClient creates a new Supervisor API client.
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) doRequest(ctx context.Context, method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			return &Error{
				StatusCode: resp.StatusCode,
				Message:    errResp.Error,
				Details:    errResp.Details,
			}
		}
		return &Error{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
		}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}

// Moderate sends a moderation request for text or image content.
func (c *Client) Moderate(ctx context.Context, req *ModerationRequest) (*ModerationResponse, error) {
	var result ModerationResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/moderate", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ModerateBatch sends a batch moderation request for multiple texts.
func (c *Client) ModerateBatch(ctx context.Context, req *BatchModerationRequest) ([]ModerationResponse, error) {
	if len(req.Texts) > 0 && len(req.Images) > 0 && len(req.Texts) != len(req.Images) {
		return nil, fmt.Errorf("texts and images must have equal length when both are provided: got %d texts and %d images", len(req.Texts), len(req.Images))
	}
	var result []ModerationResponse
	if err := c.doRequest(ctx, http.MethodPost, "/api/batch", req, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CheckUsername checks a username for policy violations.
func (c *Client) CheckUsername(ctx context.Context, username string) (*UsernameCheckResponse, error) {
	var result UsernameCheckResponse
	req := UsernameCheckRequest{Username: username}
	if err := c.doRequest(ctx, http.MethodPost, "/api/username", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLabels retrieves all available moderation labels.
func (c *Client) GetLabels(ctx context.Context) (map[string]string, error) {
	var result map[string]string
	if err := c.doRequest(ctx, http.MethodGet, "/api/labels", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
