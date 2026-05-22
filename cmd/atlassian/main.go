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
  - Jira              issues, projects, comments, worklogs, transitions (REST)
  - Confluence        pages, spaces, comments, CQL search (REST)
  - Product Discovery list JPD projects, ideas, insights (REST + GraphQL)
  - Issue links       link Jira tickets to JPD ideas and back (REST)
  - Goals             goals, metrics, status updates (GraphQL)
  - Projects          Atlas projects and goal linking (GraphQL)
  - Compass           components, scorecards, raw GraphQL (GraphQL)
  - Search            unified search across Jira and Confluence

Authentication:
  - REST commands (jira, confluence, jpd, link) need an API token.
  - GraphQL commands (goals, projects, compass, jpd insights) need OAuth.
  Run 'atlassian config init' to get started; running any command
  without credentials prints step-by-step setup instructions.

For Claude Code and other AI agents:
  Run 'atlassian docs' for a complete, always-current command reference,
  or 'atlassian docs --install' to install it as a Claude Code skill so
  the agent discovers and uses this CLI effectively.`,
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
