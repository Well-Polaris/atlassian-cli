package main

import (
	"fmt"
	"os"

	"github.com/peter/atlassian-cli/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "atlassian",
	Short:   "CLI for Atlassian APIs (Jira, Confluence, Goals, Product Discovery, Projects)",
	Version: version.Version,
	Long: `A comprehensive CLI tool for accessing Atlassian APIs.

Supports:
  - Jira (REST API)
  - Confluence (REST API)
  - Jira Product Discovery (REST + GraphQL)
  - Goals (GraphQL API)
  - Projects (GraphQL API)`,
}

func init() {
	rootCmd.SetVersionTemplate("atlassian version {{.Version}}\n")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
