package main

import (
	"fmt"
	"os"

	"github.com/peter/atlassian-cli/internal/auth"
	"github.com/peter/atlassian-cli/internal/client"
	"github.com/peter/atlassian-cli/internal/config"
)

var (
	cfg        *config.Config
	apiClient  *client.Client
	basicAuth  *auth.BasicAuth
	oauthAuth  *auth.OAuth
)

func initConfig() error {
	var err error
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	return nil
}

func initClient() error {
	if cfg == nil {
		if err := initConfig(); err != nil {
			return err
		}
	}

	if cfg.SiteURL == "" {
		return fmt.Errorf("ATLASSIAN_SITE_URL not configured. Set it in .env or environment")
	}

	// Set up Basic Auth for REST APIs
	if cfg.Email != "" && cfg.APIToken != "" {
		basicAuth = auth.NewBasicAuth(cfg.Email, cfg.APIToken)
	}

	// Set up OAuth for GraphQL APIs
	if cfg.AccessToken != "" {
		oauthAuth = auth.NewOAuth(cfg.ClientID, cfg.ClientSecret, cfg.AccessToken, cfg.RefreshToken)
	}

	apiClient = client.NewClient(cfg.SiteURL, cfg.CloudID, basicAuth, oauthAuth)

	return nil
}

func requireRESTAuth() error {
	if err := initClient(); err != nil {
		return err
	}
	if basicAuth == nil || !basicAuth.IsConfigured() {
		return fmt.Errorf("REST API authentication not configured. Set ATLASSIAN_EMAIL and ATLASSIAN_API_TOKEN")
	}
	return nil
}

func requireOAuth() error {
	if err := initClient(); err != nil {
		return err
	}
	if oauthAuth == nil || !oauthAuth.IsConfigured() {
		return fmt.Errorf("OAuth not configured. Set ATLASSIAN_ACCESS_TOKEN or run 'atlassian auth login'")
	}
	return nil
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
