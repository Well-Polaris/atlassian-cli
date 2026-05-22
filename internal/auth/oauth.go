package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Atlassian OAuth endpoints
	AuthorizationURL = "https://auth.atlassian.com/authorize"
	TokenURL         = "https://auth.atlassian.com/oauth/token"

	// DefaultScopes requests user identity plus the full Compass scope set.
	// Jira/Confluence use the API token, not OAuth, so their scopes are not
	// requested here. Override per-login with 'auth login --scopes'.
	DefaultScopes = "read:me " +
		"read:component:compass write:component:compass " +
		"read:scorecard:compass write:scorecard:compass " +
		"read:event:compass write:event:compass " +
		"read:metric:compass write:metric:compass " +
		"offline_access"
)

// OAuth provides OAuth 2.0 authentication for Atlassian GraphQL APIs
type OAuth struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Scopes       string

	// OnRefresh, if set, is called after a successful token refresh so the
	// caller can persist the rotated tokens. Atlassian rotates refresh
	// tokens, so persisting both is required for the next run to work.
	OnRefresh func(*OAuth)
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// NewOAuth creates a new OAuth instance
func NewOAuth(clientID, clientSecret, accessToken, refreshToken string) *OAuth {
	return &OAuth{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Scopes:       DefaultScopes,
	}
}

// Apply adds the Authorization header to the request
func (o *OAuth) Apply(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+o.AccessToken)
}

// IsConfigured returns true if OAuth credentials are set
func (o *OAuth) IsConfigured() bool {
	return o.AccessToken != ""
}

// HasRefreshToken returns true if a refresh token is available
func (o *OAuth) HasRefreshToken() bool {
	return o.RefreshToken != ""
}

// NeedsRefresh returns true if the token is expired or about to expire
func (o *OAuth) NeedsRefresh() bool {
	if o.ExpiresAt.IsZero() {
		return false // Unknown expiry, assume valid
	}
	// Refresh 5 minutes before expiry
	return time.Now().Add(5 * time.Minute).After(o.ExpiresAt)
}

// GetAuthorizationURL returns the URL for the OAuth authorization flow
func (o *OAuth) GetAuthorizationURL(redirectURI, state string) string {
	params := url.Values{
		"audience":      {"api.atlassian.com"},
		"client_id":     {o.ClientID},
		"scope":         {o.Scopes},
		"redirect_uri":  {redirectURI},
		"state":         {state},
		"response_type": {"code"},
		"prompt":        {"consent"},
	}
	return AuthorizationURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for tokens
func (o *OAuth) ExchangeCode(ctx context.Context, code, redirectURI string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {o.ClientID},
		"client_secret": {o.ClientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}

	return o.tokenRequest(ctx, data)
}

// Refresh refreshes the access token using the refresh token
func (o *OAuth) Refresh(ctx context.Context) (*TokenResponse, error) {
	if o.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {o.ClientID},
		"client_secret": {o.ClientSecret},
		"refresh_token": {o.RefreshToken},
	}

	resp, err := o.tokenRequest(ctx, data)
	if err != nil {
		return nil, err
	}

	// Update tokens
	o.AccessToken = resp.AccessToken
	if resp.RefreshToken != "" {
		o.RefreshToken = resp.RefreshToken
	}
	o.ExpiresAt = time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)

	if o.OnRefresh != nil {
		o.OnRefresh(o)
	}

	return resp, nil
}

func (o *OAuth) tokenRequest(ctx context.Context, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("token request failed (%d): %s - %s", resp.StatusCode, errResp.Error, errResp.Description)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	return &tokenResp, nil
}

// SaveTokens saves tokens to a file for persistence
func (o *OAuth) SaveTokens(filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	data := struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		ExpiresAt    time.Time `json:"expires_at"`
	}{
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
		ExpiresAt:    o.ExpiresAt,
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(data)
}

// LoadTokens loads tokens from a file
func (o *OAuth) LoadTokens(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	var data struct {
		AccessToken  string    `json:"access_token"`
		RefreshToken string    `json:"refresh_token"`
		ExpiresAt    time.Time `json:"expires_at"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("parsing file: %w", err)
	}

	o.AccessToken = data.AccessToken
	o.RefreshToken = data.RefreshToken
	o.ExpiresAt = data.ExpiresAt

	return nil
}
