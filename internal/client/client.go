package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/peter/atlassian-cli/internal/auth"
)

// Client is the base HTTP client for Atlassian APIs
type Client struct {
	httpClient *http.Client
	baseURL    string
	basicAuth  *auth.BasicAuth
	oauth      *auth.OAuth
	cloudID    string
}

// NewClient creates a new Atlassian API client
func NewClient(baseURL, cloudID string, basicAuth *auth.BasicAuth, oauth *auth.OAuth) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    baseURL,
		basicAuth:  basicAuth,
		oauth:      oauth,
		cloudID:    cloudID,
	}
}

// CloudID returns the configured cloud ID
func (c *Client) CloudID() string {
	return c.cloudID
}

// BaseURL returns the configured base URL
func (c *Client) BaseURL() string {
	return c.baseURL
}

// DoREST makes a REST API request with basic auth
func (c *Client) DoREST(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	return c.doRequest(ctx, method, c.baseURL+path, body, result, true)
}

// DoUpload POSTs a file as multipart/form-data using basic auth and returns
// the raw response body. Used for endpoints (e.g. Confluence attachments)
// that the JSON-only doRequest cannot serve. extraHeaders is applied last.
func (c *Client) DoUpload(ctx context.Context, path, fileField, filePath string, extraHeaders map[string]string) ([]byte, error) {
	if c.basicAuth == nil || !c.basicAuth.IsConfigured() {
		return nil, fmt.Errorf("basic auth not configured")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile(fileField, filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}
	c.basicAuth.Apply(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// DoGraphQL makes a GraphQL API request with OAuth
func (c *Client) DoGraphQL(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	if c.oauth == nil || !c.oauth.IsConfigured() {
		return fmt.Errorf("OAuth not configured - GraphQL APIs require OAuth authentication")
	}

	// Proactively refresh if we know the token is expiring.
	if c.oauth.NeedsRefresh() && c.oauth.HasRefreshToken() {
		if _, err := c.oauth.Refresh(ctx); err != nil {
			return fmt.Errorf("refreshing token: %w", err)
		}
	}

	payload := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}

	const graphqlURL = "https://api.atlassian.com/graphql"
	err := c.doRequest(ctx, "POST", graphqlURL, payload, result, false)

	// Reactive refresh: if the access token was rejected, refresh once and
	// retry. This is what makes refresh automatic across CLI invocations,
	// since the access token's expiry is not tracked between runs.
	if isUnauthorized(err) && c.oauth.HasRefreshToken() {
		if _, rerr := c.oauth.Refresh(ctx); rerr != nil {
			return fmt.Errorf("access token expired and refresh failed (%w) — run 'atlassian auth login'", rerr)
		}
		err = c.doRequest(ctx, "POST", graphqlURL, payload, result, false)
	}
	return err
}

// isUnauthorized reports whether err is a 401 from doRequest.
func isUnauthorized(err error) bool {
	return err != nil && strings.Contains(err.Error(), "API error (401)")
}

func (c *Client) doRequest(ctx context.Context, method, url string, body interface{}, result interface{}, useBasicAuth bool) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if useBasicAuth {
		if c.basicAuth == nil || !c.basicAuth.IsConfigured() {
			return fmt.Errorf("basic auth not configured")
		}
		c.basicAuth.Apply(req)
	} else {
		if c.oauth == nil || !c.oauth.IsConfigured() {
			return fmt.Errorf("OAuth not configured")
		}
		c.oauth.Apply(req)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}

	return nil
}

// GraphQLResponse wraps a GraphQL response with potential errors
type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// DoGraphQLWithErrors makes a GraphQL request and handles errors
func (c *Client) DoGraphQLWithErrors(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	var resp GraphQLResponse
	if err := c.DoGraphQL(ctx, query, variables, &resp); err != nil {
		return err
	}

	if len(resp.Errors) > 0 {
		return fmt.Errorf("GraphQL error: %s", resp.Errors[0].Message)
	}

	if result != nil {
		if err := json.Unmarshal(resp.Data, result); err != nil {
			return fmt.Errorf("parsing GraphQL data: %w", err)
		}
	}

	return nil
}
