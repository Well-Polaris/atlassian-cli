package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the Atlassian CLI
type Config struct {
	// Site configuration
	SiteURL string
	CloudID string

	// REST API auth (Basic Auth with API token)
	Email    string
	APIToken string

	// OAuth 2.0 (for GraphQL APIs)
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
}

// Load reads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try to load .env file from current directory or home
	envPaths := []string{
		".env",
		filepath.Join(os.Getenv("HOME"), ".atlassian-cli", ".env"),
	}

	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			_ = godotenv.Load(path)
			break
		}
	}

	cfg := &Config{
		SiteURL:      os.Getenv("ATLASSIAN_SITE_URL"),
		CloudID:      os.Getenv("ATLASSIAN_CLOUD_ID"),
		Email:        os.Getenv("ATLASSIAN_EMAIL"),
		APIToken:     os.Getenv("ATLASSIAN_API_TOKEN"),
		ClientID:     os.Getenv("ATLASSIAN_CLIENT_ID"),
		ClientSecret: os.Getenv("ATLASSIAN_CLIENT_SECRET"),
		AccessToken:  os.Getenv("ATLASSIAN_ACCESS_TOKEN"),
		RefreshToken: os.Getenv("ATLASSIAN_REFRESH_TOKEN"),
	}

	return cfg, nil
}

// HasRESTAuth returns true if REST API authentication is configured
func (c *Config) HasRESTAuth() bool {
	return c.Email != "" && c.APIToken != ""
}

// HasOAuth returns true if OAuth authentication is configured
func (c *Config) HasOAuth() bool {
	return c.AccessToken != ""
}
