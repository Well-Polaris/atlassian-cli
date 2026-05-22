package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration commands",
	Long:  "View and manage CLI configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initConfig(); err != nil {
			exitOnError(err)
		}

		fmt.Println("Current Configuration:")
		fmt.Println()

		// Site
		if cfg.SiteURL != "" {
			fmt.Printf("Site URL:     %s\n", cfg.SiteURL)
		} else {
			fmt.Println("Site URL:     (not set)")
		}

		if cfg.CloudID != "" {
			fmt.Printf("Cloud ID:     %s\n", cfg.CloudID)
		} else {
			fmt.Println("Cloud ID:     (not set)")
		}

		fmt.Println()

		// REST Auth
		fmt.Println("REST API Authentication (Jira/Confluence):")
		if cfg.Email != "" {
			fmt.Printf("  Email:      %s\n", cfg.Email)
		} else {
			fmt.Println("  Email:      (not set)")
		}
		if cfg.APIToken != "" {
			fmt.Println("  API Token:  ******* (set)")
		} else {
			fmt.Println("  API Token:  (not set)")
		}

		fmt.Println()

		// OAuth
		fmt.Println("OAuth Authentication (Goals/Projects GraphQL):")
		if cfg.ClientID != "" {
			fmt.Printf("  Client ID:  %s\n", cfg.ClientID)
		} else {
			fmt.Println("  Client ID:  (not set)")
		}
		if cfg.ClientSecret != "" {
			fmt.Println("  Client Secret: ******* (set)")
		} else {
			fmt.Println("  Client Secret: (not set)")
		}
		if cfg.AccessToken != "" {
			fmt.Println("  Access Token:  ******* (set)")
		} else {
			fmt.Println("  Access Token:  (not set)")
		}
		if cfg.RefreshToken != "" {
			fmt.Println("  Refresh Token: ******* (set)")
		} else {
			fmt.Println("  Refresh Token: (not set)")
		}

		fmt.Println()

		// Status
		fmt.Println("Status:")
		if cfg.HasRESTAuth() {
			fmt.Println("  ✓ REST API ready (Jira, Confluence, JPD)")
		} else {
			fmt.Println("  ✗ REST API not configured")
		}
		if cfg.HasOAuth() {
			fmt.Println("  ✓ OAuth ready (Goals, Projects, Search)")
		} else {
			fmt.Println("  ✗ OAuth not configured")
		}
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		global, _ := cmd.Flags().GetBool("global")

		var envPath string
		if global {
			homeDir, err := os.UserHomeDir()
			exitOnError(err)
			envPath = filepath.Join(homeDir, ".atlassian-cli", ".env")

			// Create directory if needed
			err = os.MkdirAll(filepath.Dir(envPath), 0700)
			exitOnError(err)
		} else {
			envPath = ".env"
		}

		// Check if file exists
		if _, err := os.Stat(envPath); err == nil {
			fmt.Printf("%s already exists. Edit it directly or use --force to overwrite.\n", envPath)
			return
		}

		template := `# Atlassian CLI Configuration
# See: https://developer.atlassian.com/cloud/

# Your Atlassian site URL
ATLASSIAN_SITE_URL=https://yoursite.atlassian.net

# Cloud ID (optional - can be auto-discovered)
# ATLASSIAN_CLOUD_ID=

# REST API Authentication (for Jira/Confluence)
# Generate at: https://id.atlassian.com/manage-profile/security/api-tokens
ATLASSIAN_EMAIL=your-email@example.com
ATLASSIAN_API_TOKEN=your-api-token

# OAuth 2.0 (for Goals/Projects GraphQL APIs)
# Create app at: https://developer.atlassian.com/console/myapps/
# ATLASSIAN_CLIENT_ID=
# ATLASSIAN_CLIENT_SECRET=
# ATLASSIAN_ACCESS_TOKEN=
# ATLASSIAN_REFRESH_TOKEN=
`

		err := os.WriteFile(envPath, []byte(template), 0600)
		exitOnError(err)

		fmt.Printf("Created %s\n", envPath)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("1. Edit the file and add your credentials")
		fmt.Println("2. For REST APIs (Jira/Confluence): set EMAIL and API_TOKEN")
		fmt.Println("3. For GraphQL APIs (Goals/Projects): set up OAuth credentials")
		fmt.Println()
		fmt.Println("Generate API token at: https://id.atlassian.com/manage-profile/security/api-tokens")
		fmt.Println("Create OAuth app at: https://developer.atlassian.com/console/myapps/")
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file paths",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, _ := os.UserHomeDir()
		globalPath := filepath.Join(homeDir, ".atlassian-cli", ".env")

		fmt.Println("Configuration file search order:")
		fmt.Println()

		// Check local
		if _, err := os.Stat(".env"); err == nil {
			fmt.Println("  1. .env (current directory) ✓ exists")
		} else {
			fmt.Println("  1. .env (current directory)")
		}

		// Check global
		if _, err := os.Stat(globalPath); err == nil {
			fmt.Printf("  2. %s ✓ exists\n", globalPath)
		} else {
			fmt.Printf("  2. %s\n", globalPath)
		}

		fmt.Println()
		fmt.Println("Environment variables can also be set directly.")
	},
}

func init() {
	configInitCmd.Flags().Bool("global", false, "Create in ~/.atlassian-cli/.env instead of current directory")
	configInitCmd.Flags().Bool("force", false, "Overwrite existing file")

	configCmd.AddCommand(configShowCmd, configInitCmd, configPathCmd)
	rootCmd.AddCommand(configCmd)
}
