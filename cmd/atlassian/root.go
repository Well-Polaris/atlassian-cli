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

// restAuthHelp explains how to obtain and configure an API token. Shown when a
// command needs REST credentials (Jira, Confluence, JPD, link) but none are set.
const restAuthHelp = `REST API authentication is not configured.

Jira, Confluence, JPD and link commands need an Atlassian API token (free):

  1. Open https://id.atlassian.com/manage-profile/security/api-tokens
  2. Click "Create API token", give it a label, set an expiry, and copy
     the token — it is shown only once.

Then configure the CLI:

  atlassian config init      # creates a .env file in the current directory

Edit the .env file and set these three values:

  ATLASSIAN_SITE_URL=https://yoursite.atlassian.net
  ATLASSIAN_EMAIL=you@example.com
  ATLASSIAN_API_TOKEN=<the token you copied>

(Or export them as environment variables.) Then re-run your command.
Check your setup any time with 'atlassian config show'.`

// oauthHelp explains how to set up OAuth, needed for the GraphQL-backed
// commands (goals, projects, jpd insights).
const oauthHelp = `OAuth authentication is not configured.

The goals, projects and 'jpd insights' commands use Atlassian's GraphQL
APIs, which require OAuth 2.0 rather than an API token:

  1. Create an OAuth 2.0 app at https://developer.atlassian.com/console/myapps/
  2. Add scopes: read:me, read:jira-work, read:confluence-content.all,
     offline_access
  3. Put ATLASSIAN_CLIENT_ID and ATLASSIAN_CLIENT_SECRET in your .env
  4. Run 'atlassian auth login' to complete the browser flow

Then re-run your command. Note: Jira, Confluence, JPD ideas and link
commands do NOT need OAuth — just an API token (see 'atlassian config init').`

func initClient() error {
	if cfg == nil {
		if err := initConfig(); err != nil {
			return err
		}
	}

	if cfg.SiteURL == "" {
		return fmt.Errorf("%s", restAuthHelp)
	}

	// Set up Basic Auth for REST APIs
	if cfg.Email != "" && cfg.APIToken != "" {
		basicAuth = auth.NewBasicAuth(cfg.Email, cfg.APIToken)
	}

	// Set up OAuth for GraphQL APIs
	if cfg.AccessToken != "" {
		oauthAuth = auth.NewOAuth(cfg.ClientID, cfg.ClientSecret, cfg.AccessToken, cfg.RefreshToken)
		// Persist rotated tokens after an automatic refresh so the next CLI
		// run starts with valid credentials.
		oauthAuth.OnRefresh = func(o *auth.OAuth) {
			_ = upsertEnv(activeEnvPath(), map[string]string{
				"ATLASSIAN_ACCESS_TOKEN":  o.AccessToken,
				"ATLASSIAN_REFRESH_TOKEN": o.RefreshToken,
			})
		}
	}

	apiClient = client.NewClient(cfg.SiteURL, cfg.CloudID, basicAuth, oauthAuth)

	return nil
}

func requireRESTAuth() error {
	if err := initClient(); err != nil {
		return err
	}
	if basicAuth == nil || !basicAuth.IsConfigured() {
		return fmt.Errorf("%s", restAuthHelp)
	}
	return nil
}

func requireOAuth() error {
	if err := initClient(); err != nil {
		return err
	}
	if oauthAuth == nil || !oauthAuth.IsConfigured() {
		return fmt.Errorf("%s", oauthHelp)
	}
	return nil
}

func exitOnError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
