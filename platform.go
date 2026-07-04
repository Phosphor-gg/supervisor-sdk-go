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

// PlatformClient is the Supervisor Platform API client with OAuth2 client credentials.
type PlatformClient struct {
	clientID     string
	clientSecret string
	baseURL      string
	httpClient   *http.Client

	mu             sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time
}

// NewPlatformClient creates a new Platform API client.
func NewPlatformClient(clientID, clientSecret string, opts ...ClientOption) *PlatformClient {
	// Use a temporary Client to apply options
	tmp := &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(tmp)
	}

	return &PlatformClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      tmp.baseURL,
		httpClient:   tmp.httpClient,
	}
}

func (p *PlatformClient) ensureToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.accessToken != "" && time.Now().Before(p.tokenExpiresAt.Add(-30*time.Second)) {
		return p.accessToken, nil
	}

	tokenReq := PlatformTokenRequest{
		ClientID:     p.clientID,
		ClientSecret: p.clientSecret,
		GrantType:    "client_credentials",
	}

	data, err := json.Marshal(tokenReq)
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/platform/token", bytes.NewReader(data))
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

	var tokenResp PlatformTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("unmarshal token response: %w", err)
	}

	p.accessToken = tokenResp.AccessToken
	p.tokenExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return p.accessToken, nil
}

func (p *PlatformClient) doRequest(ctx context.Context, method, path string, reqBody any, result any) error {
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
func (p *PlatformClient) ProvisionUser(ctx context.Context, email string) (*ProvisionUserResponse, error) {
	var result ProvisionUserResponse
	req := ProvisionUserRequest{Email: email}
	if err := p.doRequest(ctx, http.MethodPost, "/api/platform/users/provision", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListUsers lists all users linked to this platform.
func (p *PlatformClient) ListUsers(ctx context.Context) ([]PlatformUserInfo, error) {
	var result []PlatformUserInfo
	if err := p.doRequest(ctx, http.MethodGet, "/api/platform/users", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetUser gets a specific linked user by ID.
func (p *PlatformClient) GetUser(ctx context.Context, userID string) (*PlatformUserInfo, error) {
	var result PlatformUserInfo
	if err := p.doRequest(ctx, http.MethodGet, "/api/platform/users/"+userID, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Moderate moderates content on behalf of a linked user.
func (p *PlatformClient) Moderate(ctx context.Context, req *PlatformModerationRequest) (*ModerationResponse, error) {
	var result ModerationResponse
	if err := p.doRequest(ctx, http.MethodPost, "/api/platform/moderate", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCheckout creates a Stripe checkout session for a platform user.
func (p *PlatformClient) CreateCheckout(ctx context.Context, req *PlatformCheckoutRequest) (*PlatformCheckoutResponse, error) {
	var result PlatformCheckoutResponse
	if err := p.doRequest(ctx, http.MethodPost, "/api/platform/checkout", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ChangePlan changes the plan of a platform user's active subscription.
func (p *PlatformClient) ChangePlan(ctx context.Context, req PlatformChangePlanRequest) (*PlatformChangePlanResponse, error) {
	var result PlatformChangePlanResponse
	if err := p.doRequest(ctx, http.MethodPost, "/api/platform/change-plan", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetProducts lists the plans and credit packs this platform can sell.
func (p *PlatformClient) GetProducts(ctx context.Context) (*PlatformProductsResponse, error) {
	var result PlatformProductsResponse
	if err := p.doRequest(ctx, http.MethodGet, "/api/platform/products", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateCreditCheckout creates a Stripe checkout for a credit pack purchase.
func (p *PlatformClient) CreateCreditCheckout(ctx context.Context, req PlatformCreditCheckoutRequest) (*PlatformCheckoutResponse, error) {
	var result PlatformCheckoutResponse
	if err := p.doRequest(ctx, http.MethodPost, "/api/platform/checkout-credits", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetUserCredits gets the remaining credits of an authorized linked user.
func (p *PlatformClient) GetUserCredits(ctx context.Context, userID string) (*PlatformUserCreditsResponse, error) {
	var result PlatformUserCreditsResponse
	if err := p.doRequest(ctx, http.MethodGet, "/api/platform/users/"+userID+"/credits", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConfirmAuthorization confirms a user's authorization with the provided code.
func (p *PlatformClient) ConfirmAuthorization(ctx context.Context, code string) (*ConfirmAuthorizationResponse, error) {
	var result ConfirmAuthorizationResponse
	req := ConfirmAuthorizationRequest{Code: code}
	if err := p.doRequest(ctx, http.MethodPost, "/api/platform/users/confirm-authorization", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetConnectStatus gets the Stripe Connect onboarding status.
func (p *PlatformClient) GetConnectStatus(ctx context.Context) (*StripeConnectStatusResponse, error) {
	var result StripeConnectStatusResponse
	if err := p.doRequest(ctx, http.MethodGet, "/api/platform/connect/status", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
