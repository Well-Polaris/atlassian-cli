package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Well-Polaris/atlassian-cli/internal/client/projects"
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Atlas Projects commands",
	Long:  "Commands for interacting with Atlassian Projects (Atlas) GraphQL API",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := projects.New(apiClient)
		projectList, err := client.ListProjects(context.Background(), limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(projectList, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Projects (%d):\n\n", len(projectList))
			for _, p := range projectList {
				owner := ""
				if p.Owner != nil {
					owner = p.Owner.Name
				}
				fmt.Printf("%-12s [%-10s] %s (Owner: %s)\n", p.ID, p.State, p.Name, owner)
			}
		}
	},
}

var projectsGetCmd = &cobra.Command{
	Use:   "get <project-id>",
	Short: "Get a project by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		outputJSON, _ := cmd.Flags().GetBool("json")

		client := projects.New(apiClient)
		project, err := client.GetProject(context.Background(), args[0])
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(project, "", "  ")
			fmt.Println(string(data))
		} else {
			printProject(project)
		}
	},
}

var projectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		targetDate, _ := cmd.Flags().GetString("target")
		startDate, _ := cmd.Flags().GetString("start")

		input := &projects.CreateProjectInput{
			Name:        name,
			Description: description,
			TargetDate:  targetDate,
			StartDate:   startDate,
		}

		client := projects.New(apiClient)
		project, err := client.CreateProject(context.Background(), input)
		exitOnError(err)

		fmt.Printf("Created project: %s (ID: %s)\n", project.Name, project.ID)
	},
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update <project-id>",
	Short: "Update an existing project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		input := &projects.UpdateProjectInput{}

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			input.Name = &name
		}
		if cmd.Flags().Changed("description") {
			desc, _ := cmd.Flags().GetString("description")
			input.Description = &desc
		}
		if cmd.Flags().Changed("target") {
			target, _ := cmd.Flags().GetString("target")
			input.TargetDate = &target
		}
		if cmd.Flags().Changed("start") {
			start, _ := cmd.Flags().GetString("start")
			input.StartDate = &start
		}
		if cmd.Flags().Changed("state") {
			state, _ := cmd.Flags().GetString("state")
			input.State = &state
		}

		client := projects.New(apiClient)
		project, err := client.UpdateProject(context.Background(), args[0], input)
		exitOnError(err)

		fmt.Printf("Updated project: %s (ID: %s)\n", project.Name, project.ID)
	},
}

var projectsDeleteCmd = &cobra.Command{
	Use:   "delete <project-id>",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		client := projects.New(apiClient)
		err := client.DeleteProject(context.Background(), args[0])
		exitOnError(err)

		fmt.Printf("Deleted project: %s\n", args[0])
	},
}

var projectsLinkGoalCmd = &cobra.Command{
	Use:   "link-goal <project-id> <goal-id>",
	Short: "Link a goal to a project",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		client := projects.New(apiClient)
		err := client.LinkGoal(context.Background(), args[0], args[1])
		exitOnError(err)

		fmt.Printf("Linked goal %s to project %s\n", args[1], args[0])
	},
}

var projectsUnlinkGoalCmd = &cobra.Command{
	Use:   "unlink-goal <project-id> <goal-id>",
	Short: "Unlink a goal from a project",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		client := projects.New(apiClient)
		err := client.UnlinkGoal(context.Background(), args[0], args[1])
		exitOnError(err)

		fmt.Printf("Unlinked goal %s from project %s\n", args[1], args[0])
	},
}

var projectsStatusCmd = &cobra.Command{
	Use:   "status <project-id>",
	Short: "Add a status update to a project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		message, _ := cmd.Flags().GetString("message")
		state, _ := cmd.Flags().GetString("state")

		client := projects.New(apiClient)
		update, err := client.CreateProjectUpdate(context.Background(), args[0], message, state)
		exitOnError(err)

		fmt.Printf("Added status update (ID: %s) to project %s\n", update.ID, args[0])
	},
}

func printProject(project *projects.Project) {
	fmt.Printf("ID:          %s\n", project.ID)
	fmt.Printf("ARI:         %s\n", project.ARI)
	fmt.Printf("Name:        %s\n", project.Name)
	fmt.Printf("State:       %s\n", project.State)

	if project.Description != "" {
		fmt.Printf("Description: %s\n", project.Description)
	}
	if project.StartDate != "" {
		fmt.Printf("Start Date:  %s\n", project.StartDate)
	}
	if project.TargetDate != "" {
		fmt.Printf("Target Date: %s\n", project.TargetDate)
	}
	if project.Owner != nil {
		fmt.Printf("Owner:       %s\n", project.Owner.Name)
	}
	fmt.Printf("Created:     %s\n", project.CreatedAt)
	fmt.Printf("Updated:     %s\n", project.UpdatedAt)

	if len(project.Goals) > 0 {
		fmt.Printf("\nLinked Goals:\n")
		for _, g := range project.Goals {
			fmt.Printf("  - %s: %s\n", g.ID, g.Name)
		}
	}
}

func init() {
	// List command
	projectsListCmd.Flags().Int("limit", 50, "Maximum results")
	projectsListCmd.Flags().Bool("json", false, "Output as JSON")

	// Get command
	projectsGetCmd.Flags().Bool("json", false, "Output as JSON")

	// Create command
	projectsCreateCmd.Flags().String("name", "", "Project name (required)")
	projectsCreateCmd.Flags().String("description", "", "Project description")
	projectsCreateCmd.Flags().String("target", "", "Target date (ISO 8601)")
	projectsCreateCmd.Flags().String("start", "", "Start date (ISO 8601)")
	projectsCreateCmd.MarkFlagRequired("name")

	// Update command
	projectsUpdateCmd.Flags().String("name", "", "New name")
	projectsUpdateCmd.Flags().String("description", "", "New description")
	projectsUpdateCmd.Flags().String("target", "", "New target date (ISO 8601)")
	projectsUpdateCmd.Flags().String("start", "", "New start date (ISO 8601)")
	projectsUpdateCmd.Flags().String("state", "", "New state")

	// Status command
	projectsStatusCmd.Flags().String("message", "", "Update message (required)")
	projectsStatusCmd.Flags().String("state", "on_track", "State (on_track, at_risk, off_track)")
	projectsStatusCmd.MarkFlagRequired("message")

	// Add all to projects
	projectsCmd.AddCommand(projectsListCmd, projectsGetCmd, projectsCreateCmd, projectsUpdateCmd, projectsDeleteCmd, projectsLinkGoalCmd, projectsUnlinkGoalCmd, projectsStatusCmd)

	// Add to root
	rootCmd.AddCommand(projectsCmd)
}
