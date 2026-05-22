package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/peter/atlassian-cli/internal/client/goals"
	"github.com/spf13/cobra"
)

var goalsCmd = &cobra.Command{
	Use:   "goals",
	Short: "Goals commands",
	Long:  "Commands for interacting with Atlassian Goals GraphQL API",
}

var goalsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List goals",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := goals.New(apiClient)
		goalList, err := client.ListGoals(context.Background(), limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(goalList, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Goals (%d):\n\n", len(goalList))
			for _, g := range goalList {
				owner := ""
				if g.Owner != nil {
					owner = g.Owner.Name
				}
				fmt.Printf("%-12s [%-10s] %s (Owner: %s)\n", g.ID, g.State, g.Name, owner)
			}
		}
	},
}

var goalsGetCmd = &cobra.Command{
	Use:   "get <goal-id>",
	Short: "Get a goal by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		outputJSON, _ := cmd.Flags().GetBool("json")

		client := goals.New(apiClient)
		goal, err := client.GetGoal(context.Background(), args[0])
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(goal, "", "  ")
			fmt.Println(string(data))
		} else {
			printGoal(goal)
		}
	},
}

var goalsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new goal",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		dueDate, _ := cmd.Flags().GetString("due")
		startDate, _ := cmd.Flags().GetString("start")

		input := &goals.CreateGoalInput{
			Name:        name,
			Description: description,
			DueDate:     dueDate,
			StartDate:   startDate,
		}

		client := goals.New(apiClient)
		goal, err := client.CreateGoal(context.Background(), input)
		exitOnError(err)

		fmt.Printf("Created goal: %s (ID: %s)\n", goal.Name, goal.ID)
	},
}

var goalsUpdateCmd = &cobra.Command{
	Use:   "update <goal-id>",
	Short: "Update an existing goal",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		input := &goals.UpdateGoalInput{}

		if cmd.Flags().Changed("name") {
			name, _ := cmd.Flags().GetString("name")
			input.Name = &name
		}
		if cmd.Flags().Changed("description") {
			desc, _ := cmd.Flags().GetString("description")
			input.Description = &desc
		}
		if cmd.Flags().Changed("due") {
			due, _ := cmd.Flags().GetString("due")
			input.DueDate = &due
		}
		if cmd.Flags().Changed("start") {
			start, _ := cmd.Flags().GetString("start")
			input.StartDate = &start
		}
		if cmd.Flags().Changed("state") {
			state, _ := cmd.Flags().GetString("state")
			input.State = &state
		}

		client := goals.New(apiClient)
		goal, err := client.UpdateGoal(context.Background(), args[0], input)
		exitOnError(err)

		fmt.Printf("Updated goal: %s (ID: %s)\n", goal.Name, goal.ID)
	},
}

var goalsDeleteCmd = &cobra.Command{
	Use:   "delete <goal-id>",
	Short: "Delete a goal",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		client := goals.New(apiClient)
		err := client.DeleteGoal(context.Background(), args[0])
		exitOnError(err)

		fmt.Printf("Deleted goal: %s\n", args[0])
	},
}

var goalsMetricCmd = &cobra.Command{
	Use:   "metric",
	Short: "Metric commands",
}

var goalsMetricUpdateCmd = &cobra.Command{
	Use:   "update <metric-target-id> <value>",
	Short: "Update a metric target value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		var value float64
		_, err := fmt.Sscanf(args[1], "%f", &value)
		exitOnError(err)

		client := goals.New(apiClient)
		err = client.UpdateMetricValue(context.Background(), args[0], value)
		exitOnError(err)

		fmt.Printf("Updated metric %s to %v\n", args[0], value)
	},
}

var goalsUpdateCmd2 = &cobra.Command{
	Use:   "status",
	Short: "Goal status update commands",
}

var goalsStatusListCmd = &cobra.Command{
	Use:   "list <goal-id>",
	Short: "List status updates for a goal",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := goals.New(apiClient)
		updates, err := client.ListGoalUpdates(context.Background(), args[0], limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(updates, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Status updates for %s (%d):\n\n", args[0], len(updates))
			for _, u := range updates {
				author := ""
				if u.Author != nil {
					author = u.Author.Name
				}
				fmt.Printf("[%s] [%s] %s\n  %s\n\n", u.CreatedAt, u.State, author, u.Message)
			}
		}
	},
}

var goalsStatusAddCmd = &cobra.Command{
	Use:   "add <goal-id>",
	Short: "Add a status update to a goal",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireOAuth())

		message, _ := cmd.Flags().GetString("message")
		state, _ := cmd.Flags().GetString("state")

		client := goals.New(apiClient)
		update, err := client.CreateGoalUpdate(context.Background(), args[0], message, state)
		exitOnError(err)

		fmt.Printf("Added status update (ID: %s) to goal %s\n", update.ID, args[0])
	},
}

func printGoal(goal *goals.Goal) {
	fmt.Printf("ID:          %s\n", goal.ID)
	fmt.Printf("ARI:         %s\n", goal.ARI)
	fmt.Printf("Name:        %s\n", goal.Name)
	fmt.Printf("State:       %s\n", goal.State)

	if goal.Description != "" {
		fmt.Printf("Description: %s\n", goal.Description)
	}
	if goal.StartDate != "" {
		fmt.Printf("Start Date:  %s\n", goal.StartDate)
	}
	if goal.DueDate != "" {
		fmt.Printf("Due Date:    %s\n", goal.DueDate)
	}
	if goal.Owner != nil {
		fmt.Printf("Owner:       %s\n", goal.Owner.Name)
	}
	fmt.Printf("Created:     %s\n", goal.CreatedAt)
	fmt.Printf("Updated:     %s\n", goal.UpdatedAt)

	if len(goal.Metrics) > 0 {
		fmt.Printf("\nMetrics:\n")
		for _, m := range goal.Metrics {
			fmt.Printf("  - %s: %.2f / %.2f %s\n", m.Name, m.CurrentValue, m.TargetValue, m.Unit)
		}
	}
}

func init() {
	// List command
	goalsListCmd.Flags().Int("limit", 50, "Maximum results")
	goalsListCmd.Flags().Bool("json", false, "Output as JSON")

	// Get command
	goalsGetCmd.Flags().Bool("json", false, "Output as JSON")

	// Create command
	goalsCreateCmd.Flags().String("name", "", "Goal name (required)")
	goalsCreateCmd.Flags().String("description", "", "Goal description")
	goalsCreateCmd.Flags().String("due", "", "Due date (ISO 8601)")
	goalsCreateCmd.Flags().String("start", "", "Start date (ISO 8601)")
	goalsCreateCmd.MarkFlagRequired("name")

	// Update command
	goalsUpdateCmd.Flags().String("name", "", "New name")
	goalsUpdateCmd.Flags().String("description", "", "New description")
	goalsUpdateCmd.Flags().String("due", "", "New due date (ISO 8601)")
	goalsUpdateCmd.Flags().String("start", "", "New start date (ISO 8601)")
	goalsUpdateCmd.Flags().String("state", "", "New state")

	// Metric commands
	goalsMetricCmd.AddCommand(goalsMetricUpdateCmd)

	// Status commands
	goalsStatusListCmd.Flags().Int("limit", 20, "Maximum results")
	goalsStatusListCmd.Flags().Bool("json", false, "Output as JSON")
	goalsStatusAddCmd.Flags().String("message", "", "Update message (required)")
	goalsStatusAddCmd.Flags().String("state", "on_track", "State (on_track, at_risk, off_track)")
	goalsStatusAddCmd.MarkFlagRequired("message")
	goalsUpdateCmd2.AddCommand(goalsStatusListCmd, goalsStatusAddCmd)

	// Add all to goals
	goalsCmd.AddCommand(goalsListCmd, goalsGetCmd, goalsCreateCmd, goalsUpdateCmd, goalsDeleteCmd, goalsMetricCmd, goalsUpdateCmd2)

	// Add to root
	rootCmd.AddCommand(goalsCmd)
}
