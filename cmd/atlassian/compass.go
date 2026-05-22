package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/peter/atlassian-cli/internal/client/compass"
	"github.com/spf13/cobra"
)

var compassCmd = &cobra.Command{
	Use:   "compass",
	Short: "Compass commands (components, scorecards, metrics)",
	Long:  "Commands for the Compass GraphQL API. Compass is GraphQL-only and requires OAuth ('atlassian auth login').",
}

// compassCloudID returns the configured cloud ID or a clear error.
func compassCloudID() (string, error) {
	id := apiClient.CloudID()
	if id == "" {
		return "", fmt.Errorf("cloud ID not set — run 'atlassian auth login' (it resolves the cloud ID) or set ATLASSIAN_CLOUD_ID in .env")
	}
	return id, nil
}

var compassComponentCmd = &cobra.Command{
	Use:   "component",
	Short: "Compass component commands",
}

var compassComponentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List/search components in the Compass catalog",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		cloudID, err := compassCloudID()
		exitOnError(err)

		client := compass.New(apiClient)
		components, err := client.SearchComponents(context.Background(), cloudID, query, limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(components, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("Components (%d):\n\n", len(components))
		for _, c := range components {
			fmt.Printf("%-10s %s\n", c.State, c.Name)
			fmt.Printf("           %s\n", c.ID)
		}
	},
}

var compassComponentGetCmd = &cobra.Command{
	Use:   "get <component-id>",
	Short: "Get a component by its Compass ID (ARI)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		outputJSON, _ := cmd.Flags().GetBool("json")

		client := compass.New(apiClient)
		c, err := client.GetComponent(context.Background(), args[0])
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(c, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("ID:          %s\n", c.ID)
		fmt.Printf("Name:        %s\n", c.Name)
		fmt.Printf("Slug:        %s\n", c.Slug)
		fmt.Printf("State:       %s\n", c.State)
		fmt.Printf("Type ID:     %s\n", c.TypeID)
		if c.OwnerID != "" {
			fmt.Printf("Owner ID:    %s\n", c.OwnerID)
		}
		if c.URL != "" {
			fmt.Printf("URL:         %s\n", c.URL)
		}
		if c.Description != "" {
			fmt.Printf("\nDescription:\n%s\n", c.Description)
		}
	},
}

var compassScorecardCmd = &cobra.Command{
	Use:   "scorecard",
	Short: "Compass scorecard commands",
}

var compassScorecardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List scorecards in the Compass catalog",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		cloudID, err := compassCloudID()
		exitOnError(err)

		client := compass.New(apiClient)
		scorecards, err := client.ListScorecards(context.Background(), cloudID, limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(scorecards, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("Scorecards (%d):\n\n", len(scorecards))
		for _, s := range scorecards {
			fmt.Printf("%-10s %s\n", s.State, s.Name)
			fmt.Printf("           %s\n", s.ID)
		}
	},
}

var compassQueryCmd = &cobra.Command{
	Use:   "query [graphql]",
	Short: "Run a raw Compass GraphQL operation (query or mutation)",
	Long: "Run an arbitrary GraphQL operation against the Atlassian gateway. " +
		"Pass the operation as an argument or via --file. The gateway requires a " +
		"NAMED operation, e.g. 'query Name { ... }' or 'mutation Name { ... }'. " +
		"This covers everything the typed commands do not — metrics, teams, " +
		"dependencies, and all create/update mutations.",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		file, _ := cmd.Flags().GetString("file")
		varsJSON, _ := cmd.Flags().GetString("vars")

		var query string
		switch {
		case file != "":
			data, err := os.ReadFile(file)
			exitOnError(err)
			query = string(data)
		case len(args) > 0:
			query = args[0]
		default:
			exitOnError(fmt.Errorf("provide a GraphQL operation as an argument or with --file"))
		}

		var vars map[string]interface{}
		if varsJSON != "" {
			if err := json.Unmarshal([]byte(varsJSON), &vars); err != nil {
				exitOnError(fmt.Errorf("parsing --vars JSON: %w", err))
			}
		}

		client := compass.New(apiClient)
		raw, err := client.RawQuery(context.Background(), query, vars)
		exitOnError(err)

		var pretty interface{}
		if json.Unmarshal(raw, &pretty) == nil {
			out, _ := json.MarshalIndent(pretty, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Println(string(raw))
		}
	},
}

func init() {
	compassComponentListCmd.Flags().String("query", "", "Search text to filter components")
	compassComponentListCmd.Flags().Int("limit", 50, "Maximum results")
	compassComponentListCmd.Flags().Bool("json", false, "Output as JSON")
	compassComponentGetCmd.Flags().Bool("json", false, "Output as JSON")
	compassComponentCmd.AddCommand(compassComponentListCmd, compassComponentGetCmd)

	compassScorecardListCmd.Flags().Int("limit", 50, "Maximum results")
	compassScorecardListCmd.Flags().Bool("json", false, "Output as JSON")
	compassScorecardCmd.AddCommand(compassScorecardListCmd)

	compassQueryCmd.Flags().String("file", "", "Read the GraphQL operation from a file")
	compassQueryCmd.Flags().String("vars", "", "GraphQL variables as a JSON object")

	compassCmd.AddCommand(compassComponentCmd, compassScorecardCmd, compassQueryCmd)
	rootCmd.AddCommand(compassCmd)
}
