package supervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// PartnerClient is the Supervisor Partner API client with OAuth2 client credentials.
type PartnerClient struct {
	clientID     string
	clientSecret string
	baseURL      string
	httpClient   *http.Client

	mu             sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time
}

// NewPartnerClient creates a new Partner API client.
func NewPartnerClient(clientID, clientSecret string, opts ...ClientOption) *PartnerClient {
	// Use a temporary Client to apply options
	tmp := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(tmp)
	}

	return &PartnerClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      tmp.baseURL,
		httpClient:   tmp.httpClient,
	}
}

func (p *PartnerClient) ensureToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.accessToken != "" && time.Now().Before(p.tokenExpiresAt.Add(-30*time.Second)) {
		return p.accessToken, nil
	}

	tokenReq := PartnerTokenRequest{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		GrantType:    "client_credentials",
	}

	data, err := json.Marshal(tokenReq)
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/partner/token", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return "", &Error{StatusCode: resp.StatusCode, Message: errResp.Error, Details: errResp.Details}
		}
		return "", &Error{StatusCode: resp.StatusCode, Message: http.StatusText(resp.StatusCode)}
	}

	var tokenResp PartnerTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("unmarshal token response: %w", err)
	}

	p.accessToken = tokenResp.AccessToken
	p.tokenExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return p.accessToken, nil
}

func (p *PartnerClient) doRequest(ctx context.Context, method, path string, reqBody any, result any) error {
	token, err := p.ensureToken(ctx)
	if err != nil {
		return err
	}

	var body io.Reader
	if reqBody != nil {
		data, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, p.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
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
			return &Error{StatusCode: resp.StatusCode, Message: errResp.Error, Details: errResp.Details}
		}
		return &Error{StatusCode: resp.StatusCode, Message: http.StatusText(resp.StatusCode)}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}
	return nil
}

// ProvisionUser provisions or links a user by email.
func (p *PartnerClient) ProvisionUser(ctx context.Context, email string) (*ProvisionUserResponse, error) {
	var result ProvisionUserResponse
	req := ProvisionUserRequest{Email: email}
	if err := p.doRequest(ctx, http.MethodPost, "/api/partner/users/provision", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListUsers lists all users linked to this partner.
func (p *PartnerClient) ListUsers(ctx context.Context) ([]PartnerUserInfo, error) {
	var result []PartnerUserInfo
	if err := p.doRequest(ctx, http.MethodGet, "/api/partner/users", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetUser gets a specific linked user by ID.
func (p *PartnerClient) GetUser(ctx context.Context, userID string) (*PartnerUserInfo, error) {
	var result PartnerUserInfo
	if err := p.doRequest(ctx, http.MethodGet, "/api/partner/users/"+userID, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Moderate moderates content on behalf of a linked user.
func (p *PartnerClient) Moderate(ctx context.Context, req *PartnerModerationRequest) (*ModerationResponse, error) {
	var result ModerationResponse
	if err := p.doRequest(ctx, http.MethodPost, "/api/partner/moderate", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCheckout creates a Stripe checkout session for a partner user.
func (p *PartnerClient) CreateCheckout(ctx context.Context, req *PartnerCheckoutRequest) (*PartnerCheckoutResponse, error) {
	var result PartnerCheckoutResponse
	if err := p.doRequest(ctx, http.MethodPost, "/api/partner/checkout", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfirmAuthorization confirms a user's authorization with the provided code.
func (p *PartnerClient) ConfirmAuthorization(ctx context.Context, code string) (*ConfirmAuthorizationResponse, error) {
	var result ConfirmAuthorizationResponse
	req := ConfirmAuthorizationRequest{Code: code}
	if err := p.doRequest(ctx, http.MethodPost, "/api/partner/users/confirm-authorization", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetConnectStatus gets the Stripe Connect onboarding status.
func (p *PartnerClient) GetConnectStatus(ctx context.Context) (*StripeConnectStatusResponse, error) {
	var result StripeConnectStatusResponse
	if err := p.doRequest(ctx, http.MethodGet, "/api/partner/connect/status", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
