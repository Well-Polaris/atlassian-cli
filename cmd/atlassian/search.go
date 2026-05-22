package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Well-Polaris/atlassian-cli/internal/client/search"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search across Jira and Confluence",
	Long:  "Unified search using Rovo Search API",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := search.New(apiClient)
		result, err := client.Search(context.Background(), args[0], limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Search results for '%s' (%d total):\n\n", args[0], result.TotalCount)
			for _, r := range result.Results {
				container := ""
				if r.Container != nil {
					container = fmt.Sprintf(" [%s]", r.Container.Name)
				}
				fmt.Printf("[%-15s]%s %s\n", r.Type, container, r.Title)
				if r.Description != "" {
					fmt.Printf("  %s\n", truncate(r.Description, 100))
				}
				if r.URL != "" {
					fmt.Printf("  %s\n", r.URL)
				}
				fmt.Println()
			}
			if result.HasMore {
				fmt.Println("... more results available")
			}
		}
	},
}

var fetchCmd = &cobra.Command{
	Use:   "fetch <ari>",
	Short: "Fetch a resource by ARI",
	Long:  "Fetch details of a Jira issue, Confluence page, Goal, or Project by its Atlassian Resource Identifier",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		client := search.New(apiClient)
		result, err := client.FetchByARI(context.Background(), args[0])
		exitOnError(err)

		// Pretty print the JSON
		var pretty map[string]interface{}
		json.Unmarshal(result, &pretty)
		data, _ := json.MarshalIndent(pretty, "", "  ")
		fmt.Println(string(data))
	},
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	searchCmd.Flags().Int("limit", 25, "Maximum results")
	searchCmd.Flags().Bool("json", false, "Output as JSON")

	rootCmd.AddCommand(searchCmd, fetchCmd)
}
