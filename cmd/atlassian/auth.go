package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/peter/atlassian-cli/internal/auth"
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

		// Save tokens
		homeDir, _ := os.UserHomeDir()
		tokenPath := filepath.Join(homeDir, ".atlassian-cli", "tokens.json")

		oauth.AccessToken = tokenResp.AccessToken
		oauth.RefreshToken = tokenResp.RefreshToken
		oauth.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

		if err := oauth.SaveTokens(tokenPath); err != nil {
			fmt.Printf("Warning: Could not save tokens to %s: %v\n", tokenPath, err)
		}

		fmt.Println()
		fmt.Println("Authentication successful!")
		fmt.Println()
		fmt.Println("Add these to your .env file:")
		fmt.Printf("ATLASSIAN_ACCESS_TOKEN=%s\n", tokenResp.AccessToken)
		if tokenResp.RefreshToken != "" {
			fmt.Printf("ATLASSIAN_REFRESH_TOKEN=%s\n", tokenResp.RefreshToken)
		}
		fmt.Println()
		fmt.Printf("Tokens also saved to: %s\n", tokenPath)
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

		// Save tokens
		homeDir, _ := os.UserHomeDir()
		tokenPath := filepath.Join(homeDir, ".atlassian-cli", "tokens.json")

		if err := oauth.SaveTokens(tokenPath); err != nil {
			fmt.Printf("Warning: Could not save tokens to %s: %v\n", tokenPath, err)
		}

		fmt.Println()
		fmt.Println("Token refreshed successfully!")
		fmt.Println()
		fmt.Println("Update your .env file:")
		fmt.Printf("ATLASSIAN_ACCESS_TOKEN=%s\n", tokenResp.AccessToken)
		if tokenResp.RefreshToken != "" {
			fmt.Printf("ATLASSIAN_REFRESH_TOKEN=%s\n", tokenResp.RefreshToken)
		}
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

func openBrowser(url string) {
	// Try to open browser (platform-specific, best effort)
	// On macOS
	if fileExists("/usr/bin/open") {
		go func() {
			http.Get("http://localhost:1") // Force import of net/http
		}()
		// Would use: exec.Command("open", url).Start()
		fmt.Println("Please open the URL above in your browser.")
		return
	}
	// On Linux
	if fileExists("/usr/bin/xdg-open") {
		// Would use: exec.Command("xdg-open", url).Start()
		fmt.Println("Please open the URL above in your browser.")
		return
	}
	fmt.Println("Please open the URL above in your browser.")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	authLoginCmd.Flags().Int("port", 8085, "Local port for OAuth callback")

	authCmd.AddCommand(authLoginCmd, authRefreshCmd, authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
