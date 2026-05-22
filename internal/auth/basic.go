package auth

import (
	"encoding/base64"
	"net/http"
)

// BasicAuth provides HTTP Basic Authentication for Atlassian REST APIs
type BasicAuth struct {
	Email    string
	APIToken string
}

// NewBasicAuth creates a new BasicAuth instance
func NewBasicAuth(email, apiToken string) *BasicAuth {
	return &BasicAuth{
		Email:    email,
		APIToken: apiToken,
	}
}

// Apply adds the Authorization header to the request
func (b *BasicAuth) Apply(req *http.Request) {
	credentials := base64.StdEncoding.EncodeToString([]byte(b.Email + ":" + b.APIToken))
	req.Header.Set("Authorization", "Basic "+credentials)
}

// IsConfigured returns true if credentials are set
func (b *BasicAuth) IsConfigured() bool {
	return b.Email != "" && b.APIToken != ""
}
