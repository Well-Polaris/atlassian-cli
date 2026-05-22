package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/Well-Polaris/atlassian-cli/internal/client/compass"
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
		printComponent(c)
	},
}

// printComponent renders a component as a human-readable detail block.
func printComponent(c *compass.Component) {
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
}

var compassComponentCreateCmd = &cobra.Command{
	Use:   "create --name <name>",
	Short: "Create a new component in the Compass catalog",
	Long: "Create a Compass component. --name is required; --type defaults to SERVICE.\n\n" +
		"Common type IDs: SERVICE, LIBRARY, APPLICATION, CAPABILITY, WEBSITE,\n" +
		"CLOUD_RESOURCE, DATA_PIPELINE, MACHINE_LEARNING_MODEL, UI_ELEMENT, OTHER.\n" +
		"--owner takes a team ARI (ari:cloud:identity::team/...).",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		cloudID, err := compassCloudID()
		exitOnError(err)

		name, _ := cmd.Flags().GetString("name")
		typeID, _ := cmd.Flags().GetString("type")
		description, _ := cmd.Flags().GetString("description")
		slug, _ := cmd.Flags().GetString("slug")
		owner, _ := cmd.Flags().GetString("owner")
		state, _ := cmd.Flags().GetString("state")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := compass.New(apiClient)
		c, err := client.CreateComponent(context.Background(), cloudID, compass.CreateComponentInput{
			Name:        name,
			TypeID:      typeID,
			Description: description,
			Slug:        slug,
			OwnerID:     owner,
			State:       state,
		})
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(c, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("Created component: %s\n\n", c.Name)
		printComponent(c)
	},
}

var compassComponentUpdateCmd = &cobra.Command{
	Use:   "update <component-id>",
	Short: "Update an existing component",
	Long: "Update a component's editable fields. Only the flags you pass are changed; " +
		"everything else is left as-is.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		// Collect only the flags the user actually set, mapped to their
		// GraphQL input field names.
		changes := map[string]interface{}{}
		for flag, field := range map[string]string{
			"name":        "name",
			"description": "description",
			"slug":        "slug",
			"state":       "state",
			"owner":       "ownerId",
		} {
			if cmd.Flags().Changed(flag) {
				v, _ := cmd.Flags().GetString(flag)
				changes[field] = v
			}
		}
		if len(changes) == 0 {
			exitOnError(fmt.Errorf("nothing to update — pass at least one of --name, --description, --slug, --state, --owner"))
		}

		outputJSON, _ := cmd.Flags().GetBool("json")

		client := compass.New(apiClient)
		c, err := client.UpdateComponent(context.Background(), args[0], changes)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(c, "", "  ")
			fmt.Println(string(data))
			return
		}
		fmt.Printf("Updated component: %s\n\n", c.Name)
		printComponent(c)
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

	compassComponentCreateCmd.Flags().String("name", "", "Component name (required)")
	compassComponentCreateCmd.Flags().String("type", "SERVICE", "Component type ID")
	compassComponentCreateCmd.Flags().String("description", "", "Component description")
	compassComponentCreateCmd.Flags().String("slug", "", "Component slug")
	compassComponentCreateCmd.Flags().String("owner", "", "Owning team ARI")
	compassComponentCreateCmd.Flags().String("state", "", "Lifecycle state (e.g. ACTIVE)")
	compassComponentCreateCmd.Flags().Bool("json", false, "Output as JSON")
	compassComponentCreateCmd.MarkFlagRequired("name")

	compassComponentUpdateCmd.Flags().String("name", "", "New name")
	compassComponentUpdateCmd.Flags().String("description", "", "New description")
	compassComponentUpdateCmd.Flags().String("slug", "", "New slug")
	compassComponentUpdateCmd.Flags().String("owner", "", "New owning team ARI")
	compassComponentUpdateCmd.Flags().String("state", "", "New lifecycle state")
	compassComponentUpdateCmd.Flags().Bool("json", false, "Output as JSON")

	compassComponentCmd.AddCommand(
		compassComponentListCmd, compassComponentGetCmd,
		compassComponentCreateCmd, compassComponentUpdateCmd,
	)

	compassScorecardListCmd.Flags().Int("limit", 50, "Maximum results")
	compassScorecardListCmd.Flags().Bool("json", false, "Output as JSON")
	compassScorecardCmd.AddCommand(compassScorecardListCmd)

	compassQueryCmd.Flags().String("file", "", "Read the GraphQL operation from a file")
	compassQueryCmd.Flags().String("vars", "", "GraphQL variables as a JSON object")

	compassCmd.AddCommand(compassComponentCmd, compassScorecardCmd, compassQueryCmd)
	rootCmd.AddCommand(compassCmd)
}
