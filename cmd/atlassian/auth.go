package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Well-Polaris/atlassian-cli/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Manage OAuth authentication for GraphQL APIs",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with OAuth",
	Long: `Start OAuth flow to get access tokens for GraphQL APIs (Goals, Projects).

Requires ATLASSIAN_CLIENT_ID and ATLASSIAN_CLIENT_SECRET to be set.
Create an OAuth app at: https://developer.atlassian.com/console/myapps/`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initConfig(); err != nil {
			exitOnError(err)
		}

		if cfg.ClientID == "" || cfg.ClientSecret == "" {
			exitOnError(fmt.Errorf("ATLASSIAN_CLIENT_ID and ATLASSIAN_CLIENT_SECRET must be set"))
		}

		port, _ := cmd.Flags().GetInt("port")
		redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)

		// Generate state for CSRF protection
		stateBytes := make([]byte, 16)
		rand.Read(stateBytes)
		state := hex.EncodeToString(stateBytes)

		oauth := auth.NewOAuth(cfg.ClientID, cfg.ClientSecret, "", "")
		if scopes, _ := cmd.Flags().GetString("scopes"); scopes != "" {
			oauth.Scopes = scopes
		}
		authURL := oauth.GetAuthorizationURL(redirectURI, state)

		fmt.Println("Opening browser for authentication...")
		fmt.Printf("\nIf browser doesn't open, visit:\n%s\n\n", authURL)

		// Start local server to receive callback
		codeChan := make(chan string)
		errChan := make(chan error)

		server := &http.Server{Addr: fmt.Sprintf(":%d", port)}

		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			receivedState := r.URL.Query().Get("state")
			if receivedState != state {
				http.Error(w, "Invalid state", http.StatusBadRequest)
				errChan <- fmt.Errorf("state mismatch")
				return
			}

			code := r.URL.Query().Get("code")
			if code == "" {
				errMsg := r.URL.Query().Get("error_description")
				if errMsg == "" {
					errMsg = r.URL.Query().Get("error")
				}
				http.Error(w, errMsg, http.StatusBadRequest)
				errChan <- fmt.Errorf("authorization failed: %s", errMsg)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `<html><body><h1>Authentication successful!</h1><p>You can close this window.</p></body></html>`)
			codeChan <- code
		})

		go func() {
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				errChan <- err
			}
		}()

		// Open browser (best effort)
		openBrowser(authURL)

		// Wait for callback
		var code string
		select {
		case code = <-codeChan:
		case err := <-errChan:
			server.Shutdown(context.Background())
			exitOnError(err)
		case <-time.After(5 * time.Minute):
			server.Shutdown(context.Background())
			exitOnError(fmt.Errorf("authentication timed out"))
		}

		server.Shutdown(context.Background())

		// Exchange code for tokens
		fmt.Println("Exchanging code for tokens...")
		tokenResp, err := oauth.ExchangeCode(context.Background(), code, redirectURI)
		exitOnError(err)

		oauth.AccessToken = tokenResp.AccessToken
		oauth.RefreshToken = tokenResp.RefreshToken
		oauth.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

		// Save a copy to ~/.atlassian-cli/tokens.json (best effort).
		homeDir, _ := os.UserHomeDir()
		tokenPath := filepath.Join(homeDir, ".atlassian-cli", "tokens.json")
		if err := oauth.SaveTokens(tokenPath); err != nil {
			fmt.Printf("Warning: could not save tokens to %s: %v\n", tokenPath, err)
		}

		// Write tokens (and the cloud ID, needed for GraphQL/Compass) into
		// the .env the CLI actually reads, so no manual copying is needed.
		env := map[string]string{"ATLASSIAN_ACCESS_TOKEN": tokenResp.AccessToken}
		if tokenResp.RefreshToken != "" {
			env["ATLASSIAN_REFRESH_TOKEN"] = tokenResp.RefreshToken
		}
		if cloudID, err := fetchCloudID(context.Background(), tokenResp.AccessToken, cfg.SiteURL); err != nil {
			fmt.Printf("Warning: could not resolve cloud ID: %v\n", err)
		} else if cloudID != "" {
			env["ATLASSIAN_CLOUD_ID"] = cloudID
		}

		envPath := activeEnvPath()
		if err := upsertEnv(envPath, env); err != nil {
			fmt.Printf("Warning: could not update %s: %v\n", envPath, err)
		}

		fmt.Println()
		fmt.Println("Authentication successful.")
		fmt.Printf("Updated %s with access token, refresh token and cloud ID.\n", envPath)
		fmt.Printf("Granted scopes: %s\n", tokenResp.Scope)
	},
}

var authRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh access token",
	Long:  "Use the refresh token to get a new access token",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initConfig(); err != nil {
			exitOnError(err)
		}

		if cfg.RefreshToken == "" {
			exitOnError(fmt.Errorf("ATLASSIAN_REFRESH_TOKEN must be set"))
		}

		oauth := auth.NewOAuth(cfg.ClientID, cfg.ClientSecret, cfg.AccessToken, cfg.RefreshToken)

		fmt.Println("Refreshing access token...")
		tokenResp, err := oauth.Refresh(context.Background())
		exitOnError(err)

		// Save a copy to ~/.atlassian-cli/tokens.json (best effort).
		homeDir, _ := os.UserHomeDir()
		tokenPath := filepath.Join(homeDir, ".atlassian-cli", "tokens.json")
		if err := oauth.SaveTokens(tokenPath); err != nil {
			fmt.Printf("Warning: could not save tokens to %s: %v\n", tokenPath, err)
		}

		env := map[string]string{"ATLASSIAN_ACCESS_TOKEN": tokenResp.AccessToken}
		if tokenResp.RefreshToken != "" {
			env["ATLASSIAN_REFRESH_TOKEN"] = tokenResp.RefreshToken
		}
		envPath := activeEnvPath()
		if err := upsertEnv(envPath, env); err != nil {
			fmt.Printf("Warning: could not update %s: %v\n", envPath, err)
		}

		fmt.Println()
		fmt.Printf("Token refreshed. Updated %s.\n", envPath)
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initConfig(); err != nil {
			exitOnError(err)
		}

		fmt.Println("Authentication Status:")
		fmt.Println()

		// REST API
		if cfg.HasRESTAuth() {
			fmt.Println("REST API (Jira/Confluence/JPD):")
			fmt.Printf("  ✓ Configured for %s\n", cfg.Email)
		} else {
			fmt.Println("REST API (Jira/Confluence/JPD):")
			fmt.Println("  ✗ Not configured")
			fmt.Println("  Set ATLASSIAN_EMAIL and ATLASSIAN_API_TOKEN")
		}

		fmt.Println()

		// OAuth
		if cfg.HasOAuth() {
			fmt.Println("OAuth (Goals/Projects/Search):")
			fmt.Println("  ✓ Access token configured")
			if cfg.RefreshToken != "" {
				fmt.Println("  ✓ Refresh token available")
			}
		} else {
			fmt.Println("OAuth (Goals/Projects/Search):")
			fmt.Println("  ✗ Not configured")
			fmt.Println("  Run 'atlassian auth login' to authenticate")
		}
	},
}

// openBrowser opens url in the default browser (best effort).
func openBrowser(url string) {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", url)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		c = exec.Command("xdg-open", url)
	}
	if err := c.Start(); err != nil {
		fmt.Println("Could not open a browser automatically — open the URL above manually.")
	}
}

// activeEnvPath returns the .env file the CLI reads — the one in the current
// directory if present, otherwise the global ~/.atlassian-cli/.env, defaulting
// to ./.env when neither exists yet.
func activeEnvPath() string {
	if _, err := os.Stat(".env"); err == nil {
		return ".env"
	}
	home, _ := os.UserHomeDir()
	global := filepath.Join(home, ".atlassian-cli", ".env")
	if _, err := os.Stat(global); err == nil {
		return global
	}
	return ".env"
}

// upsertEnv sets KEY=value pairs in a .env file, replacing existing keys in
// place and appending any that are new. Comments and other lines are kept.
func upsertEnv(path string, kv map[string]string) error {
	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		lines = strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}

	seen := map[string]bool{}
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		eq := strings.Index(trimmed, "=")
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:eq])
		if v, ok := kv[key]; ok {
			lines[i] = key + "=" + v
			seen[key] = true
		}
	}

	newKeys := make([]string, 0, len(kv))
	for k := range kv {
		if !seen[k] {
			newKeys = append(newKeys, k)
		}
	}
	sort.Strings(newKeys)
	for _, k := range newKeys {
		lines = append(lines, k+"="+kv[k])
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}

// fetchCloudID resolves the Atlassian cloud (site) ID for the authenticated
// user, preferring the resource whose URL matches siteURL.
func fetchCloudID(ctx context.Context, accessToken, siteURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://api.atlassian.com/oauth/token/accessible-resources", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("accessible-resources returned %d: %s", resp.StatusCode, string(body))
	}

	var resources []struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &resources); err != nil {
		return "", err
	}

	want := strings.TrimSuffix(strings.TrimSpace(siteURL), "/")
	for _, r := range resources {
		if strings.EqualFold(strings.TrimSuffix(r.URL, "/"), want) {
			return r.ID, nil
		}
	}
	if len(resources) > 0 {
		return resources[0].ID, nil
	}
	return "", nil
}

func init() {
	authLoginCmd.Flags().Int("port", 8085, "Local port for OAuth callback")
	authLoginCmd.Flags().String("scopes", "", "Override OAuth scopes (space-separated); defaults to all supported APIs")

	authCmd.AddCommand(authLoginCmd, authRefreshCmd, authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
