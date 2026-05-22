package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Well-Polaris/atlassian-cli/internal/client/jira"
	"github.com/Well-Polaris/atlassian-cli/internal/client/jpd"
	"github.com/spf13/cobra"
)

var jpdCmd = &cobra.Command{
	Use:   "jpd",
	Short: "Jira Product Discovery commands",
	Long:  "Commands for interacting with Jira Product Discovery (JPD)",
}

var jpdIdeasCmd = &cobra.Command{
	Use:   "ideas",
	Short: "Idea commands",
}

var jpdIdeasListCmd = &cobra.Command{
	Use:   "list --project <project-key>",
	Short: "List ideas in a project",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		project, _ := cmd.Flags().GetString("project")
		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := jpd.New(apiClient)
		ideas, err := client.ListIdeas(context.Background(), project, limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(ideas, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Ideas in %s (%d):\n\n", project, len(ideas))
			for _, i := range ideas {
				fmt.Printf("%-12s [%-12s] %s\n", i.Key, i.Status, i.Summary)
			}
		}
	},
}

var jpdIdeasGetCmd = &cobra.Command{
	Use:   "get <idea-key>",
	Short: "Get an idea by key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		outputJSON, _ := cmd.Flags().GetBool("json")

		client := jpd.New(apiClient)
		idea, err := client.GetIdea(context.Background(), args[0])
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(idea, "", "  ")
			fmt.Println(string(data))
		} else {
			printIdea(idea)
		}
	},
}

var jpdIdeasCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new idea",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		project, _ := cmd.Flags().GetString("project")
		summary, _ := cmd.Flags().GetString("summary")
		description, _ := cmd.Flags().GetString("description")
		labels, _ := cmd.Flags().GetStringSlice("labels")

		input := &jpd.CreateIdeaInput{
			ProjectKey:  project,
			Summary:     summary,
			Description: description,
			Labels:      labels,
		}

		client := jpd.New(apiClient)
		idea, err := client.CreateIdea(context.Background(), input)
		exitOnError(err)

		fmt.Printf("Created idea: %s\n", idea.Key)
	},
}

var jpdIdeasUpdateCmd = &cobra.Command{
	Use:   "update <idea-key>",
	Short: "Update an existing idea",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		input := &jpd.UpdateIdeaInput{}

		if cmd.Flags().Changed("summary") {
			summary, _ := cmd.Flags().GetString("summary")
			input.Summary = &summary
		}
		if cmd.Flags().Changed("description") {
			desc, _ := cmd.Flags().GetString("description")
			input.Description = &desc
		}
		if cmd.Flags().Changed("labels") {
			labels, _ := cmd.Flags().GetStringSlice("labels")
			input.Labels = labels
		}

		client := jpd.New(apiClient)
		err := client.UpdateIdea(context.Background(), args[0], input)
		exitOnError(err)

		fmt.Printf("Updated idea: %s\n", args[0])
	},
}

var jpdIdeasDeleteCmd = &cobra.Command{
	Use:   "delete <idea-key>",
	Short: "Delete an idea",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jpd.New(apiClient)
		err := client.DeleteIdea(context.Background(), args[0])
		exitOnError(err)

		fmt.Printf("Deleted idea: %s\n", args[0])
	},
}

var jpdIdeasFieldsCmd = &cobra.Command{
	Use:   "fields <idea-key>",
	Short: "List all editable fields on an idea (field ID, name, type)",
	Long: "List every field that can be set on a JPD idea — including project-specific\n" +
		"custom fields like target dates, ratings and selects. JPD fields are\n" +
		"configured per project, so this reads the given idea's project.\n\n" +
		"Use the field IDs shown here with future field-setting commands.",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		outputJSON, _ := cmd.Flags().GetBool("json")

		client := jira.New(apiClient)
		fields, err := client.GetEditableFields(context.Background(), args[0])
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(fields, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Printf("Editable fields on %s (%d):\n\n", args[0], len(fields))
		fmt.Printf("%-22s %-28s %s\n", "FIELD ID", "NAME", "TYPE")
		for _, f := range fields {
			typ := f.Type
			if f.Custom != "" {
				typ += " / " + f.Custom
			}
			if f.Required {
				typ += " (required)"
			}
			fmt.Printf("%-22s %-28s %s\n", f.ID, f.Name, typ)
		}
	},
}

var jpdProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Product Discovery project commands",
}

var jpdProjectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Jira Product Discovery projects",
	Long:  "List every JPD (product_discovery) project visible to the account — the project access the Atlassian MCP omits.",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		outputJSON, _ := cmd.Flags().GetBool("json")

		client := jira.New(apiClient)
		projects, err := client.SearchProjects(context.Background(), "product_discovery")
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(projects, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Product Discovery projects (%d):\n\n", len(projects))
			for _, p := range projects {
				fmt.Printf("%-10s %s\n", p.Key, p.Name)
			}
		}
	},
}

var jpdInsightsCmd = &cobra.Command{
	Use:   "insights",
	Short: "Insight commands",
}

var jpdInsightsListCmd = &cobra.Command{
	Use:   "list <idea-key>",
	Short: "List insights for an idea",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		outputJSON, _ := cmd.Flags().GetBool("json")

		client := jpd.New(apiClient)
		insights, err := client.GetInsights(context.Background(), args[0])
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(insights, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Insights for %s (%d):\n\n", args[0], len(insights))
			for _, i := range insights {
				fmt.Printf("[%s] %s\n", i.CreatedAt, i.Description)
				if i.Source != "" {
					fmt.Printf("  Source: %s\n", i.Source)
				}
				fmt.Println()
			}
		}
	},
}

var jpdCommentCmd = &cobra.Command{
	Use:   "comment <idea-key> <comment>",
	Short: "Add a comment to an idea",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jpd.New(apiClient)
		err := client.AddComment(context.Background(), args[0], args[1])
		exitOnError(err)

		fmt.Printf("Added comment to %s\n", args[0])
	},
}

func printIdea(idea *jpd.Idea) {
	fmt.Printf("Key:         %s\n", idea.Key)
	fmt.Printf("Summary:     %s\n", idea.Summary)
	fmt.Printf("Status:      %s\n", idea.Status)
	fmt.Printf("Project:     %s\n", idea.ProjectKey)

	if idea.IdeaType != "" {
		fmt.Printf("Type:        %s\n", idea.IdeaType)
	}
	if idea.Priority != "" {
		fmt.Printf("Priority:    %s\n", idea.Priority)
	}
	if len(idea.Labels) > 0 {
		fmt.Printf("Labels:      %v\n", idea.Labels)
	}

	fmt.Printf("Created:     %s\n", idea.Created)
	fmt.Printf("Updated:     %s\n", idea.Updated)

	if idea.Description != "" {
		fmt.Printf("\nDescription:\n%s\n", idea.Description)
	}
}

func init() {
	// Ideas list command
	jpdIdeasListCmd.Flags().StringP("project", "p", "", "Project key (required)")
	jpdIdeasListCmd.Flags().Int("limit", 50, "Maximum results")
	jpdIdeasListCmd.Flags().Bool("json", false, "Output as JSON")
	jpdIdeasListCmd.MarkFlagRequired("project")

	// Ideas get command
	jpdIdeasGetCmd.Flags().Bool("json", false, "Output as JSON")

	// Ideas create command
	jpdIdeasCreateCmd.Flags().StringP("project", "p", "", "Project key (required)")
	jpdIdeasCreateCmd.Flags().StringP("summary", "s", "", "Idea summary (required)")
	jpdIdeasCreateCmd.Flags().StringP("description", "d", "", "Idea description")
	jpdIdeasCreateCmd.Flags().StringSlice("labels", nil, "Labels")
	jpdIdeasCreateCmd.MarkFlagRequired("project")
	jpdIdeasCreateCmd.MarkFlagRequired("summary")

	// Ideas update command
	jpdIdeasUpdateCmd.Flags().StringP("summary", "s", "", "New summary")
	jpdIdeasUpdateCmd.Flags().StringP("description", "d", "", "New description")
	jpdIdeasUpdateCmd.Flags().StringSlice("labels", nil, "New labels")

	// Ideas fields command
	jpdIdeasFieldsCmd.Flags().Bool("json", false, "Output as JSON")

	jpdIdeasCmd.AddCommand(jpdIdeasListCmd, jpdIdeasGetCmd, jpdIdeasCreateCmd, jpdIdeasUpdateCmd, jpdIdeasDeleteCmd, jpdIdeasFieldsCmd)

	// Projects command
	jpdProjectsListCmd.Flags().Bool("json", false, "Output as JSON")
	jpdProjectsCmd.AddCommand(jpdProjectsListCmd)

	// Insights command
	jpdInsightsListCmd.Flags().Bool("json", false, "Output as JSON")
	jpdInsightsCmd.AddCommand(jpdInsightsListCmd)

	// Add all to jpd
	jpdCmd.AddCommand(jpdIdeasCmd, jpdProjectsCmd, jpdInsightsCmd, jpdCommentCmd)

	// Add to root
	rootCmd.AddCommand(jpdCmd)
}
