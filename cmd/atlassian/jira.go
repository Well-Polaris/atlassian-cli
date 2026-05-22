package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/peter/atlassian-cli/internal/client/jira"
	"github.com/spf13/cobra"
)

var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Jira commands",
	Long:  "Commands for interacting with Jira REST API",
}

var jiraIssueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Issue commands",
}

var jiraIssueGetCmd = &cobra.Command{
	Use:   "get <issue-key>",
	Short: "Get an issue by key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)
		issue, err := client.GetIssue(context.Background(), args[0])
		exitOnError(err)

		outputJSON, _ := cmd.Flags().GetBool("json")
		if outputJSON {
			data, _ := json.MarshalIndent(issue, "", "  ")
			fmt.Println(string(data))
		} else {
			printIssue(issue)
		}
	},
}

var jiraIssueCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new issue",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		project, _ := cmd.Flags().GetString("project")
		issueType, _ := cmd.Flags().GetString("type")
		summary, _ := cmd.Flags().GetString("summary")
		description, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetString("priority")
		labels, _ := cmd.Flags().GetStringSlice("labels")
		parent, _ := cmd.Flags().GetString("parent")

		input := &jira.CreateIssueInput{
			ProjectKey:  project,
			IssueType:   issueType,
			Summary:     summary,
			Description: description,
			Priority:    priority,
			Labels:      labels,
			ParentKey:   parent,
		}

		client := jira.New(apiClient)
		issue, err := client.CreateIssue(context.Background(), input)
		exitOnError(err)

		fmt.Printf("Created issue: %s\n", issue.Key)
	},
}

var jiraIssueUpdateCmd = &cobra.Command{
	Use:   "update <issue-key>",
	Short: "Update an existing issue",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		input := &jira.UpdateIssueInput{}

		if cmd.Flags().Changed("summary") {
			summary, _ := cmd.Flags().GetString("summary")
			input.Summary = &summary
		}
		if cmd.Flags().Changed("description") {
			desc, _ := cmd.Flags().GetString("description")
			input.Description = &desc
		}
		if cmd.Flags().Changed("priority") {
			priority, _ := cmd.Flags().GetString("priority")
			input.Priority = &priority
		}
		if cmd.Flags().Changed("labels") {
			labels, _ := cmd.Flags().GetStringSlice("labels")
			input.Labels = labels
		}

		client := jira.New(apiClient)
		err := client.UpdateIssue(context.Background(), args[0], input)
		exitOnError(err)

		fmt.Printf("Updated issue: %s\n", args[0])
	},
}

var jiraIssueDeleteCmd = &cobra.Command{
	Use:   "delete <issue-key>",
	Short: "Delete an issue",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)
		err := client.DeleteIssue(context.Background(), args[0])
		exitOnError(err)

		fmt.Printf("Deleted issue: %s\n", args[0])
	},
}

var jiraIssueSearchCmd = &cobra.Command{
	Use:   "search <jql>",
	Short: "Search issues using JQL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		maxResults, _ := cmd.Flags().GetInt("max")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := jira.New(apiClient)
		result, err := client.SearchIssues(context.Background(), args[0], maxResults)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Found %d issues:\n\n", len(result.Issues))
			for _, issue := range result.Issues {
				status := ""
				if issue.Fields.Status != nil {
					status = issue.Fields.Status.Name
				}
				fmt.Printf("%-12s [%-12s] %s\n", issue.Key, status, issue.Fields.Summary)
			}
		}
	},
}

var jiraIssueTransitionCmd = &cobra.Command{
	Use:   "transition <issue-key> [status]",
	Short: "Transition an issue to a new status",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)

		// Get available transitions
		transitions, err := client.GetTransitions(context.Background(), args[0])
		exitOnError(err)

		if len(args) == 1 {
			// Just list transitions
			fmt.Printf("Available transitions for %s:\n", args[0])
			for _, t := range transitions {
				fmt.Printf("  %s (ID: %s) -> %s\n", t.Name, t.ID, t.To.Name)
			}
			return
		}

		// Find matching transition
		targetStatus := args[1]
		var transitionID string
		for _, t := range transitions {
			if t.Name == targetStatus || t.To.Name == targetStatus || t.ID == targetStatus {
				transitionID = t.ID
				break
			}
		}

		if transitionID == "" {
			exitOnError(fmt.Errorf("no transition found for status: %s", targetStatus))
		}

		err = client.TransitionIssue(context.Background(), args[0], transitionID)
		exitOnError(err)

		fmt.Printf("Transitioned %s to %s\n", args[0], targetStatus)
	},
}

var jiraCommentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Comment commands",
}

var jiraCommentListCmd = &cobra.Command{
	Use:   "list <issue-key>",
	Short: "List comments on an issue",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)
		comments, err := client.GetComments(context.Background(), args[0])
		exitOnError(err)

		outputJSON, _ := cmd.Flags().GetBool("json")
		if outputJSON {
			data, _ := json.MarshalIndent(comments, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Comments on %s (%d):\n\n", args[0], len(comments))
			for _, c := range comments {
				author := "Unknown"
				if c.Author != nil {
					author = c.Author.DisplayName
				}
				fmt.Printf("[%s] %s:\n%v\n\n", c.Created, author, c.Body)
			}
		}
	},
}

var jiraCommentAddCmd = &cobra.Command{
	Use:   "add <issue-key> <comment>",
	Short: "Add a comment to an issue",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)
		comment, err := client.AddComment(context.Background(), args[0], args[1])
		exitOnError(err)

		fmt.Printf("Added comment (ID: %s) to %s\n", comment.ID, args[0])
	},
}

var jiraProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project commands",
}

var jiraProjectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Long:  "List projects. Use --type to filter by project type, e.g. --type product_discovery for JPD projects.",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		typeKey, _ := cmd.Flags().GetString("type")
		client := jira.New(apiClient)

		var projects []*jira.Project
		var err error
		if typeKey != "" {
			projects, err = client.SearchProjects(context.Background(), typeKey)
		} else {
			projects, err = client.ListProjects(context.Background())
		}
		exitOnError(err)

		outputJSON, _ := cmd.Flags().GetBool("json")
		if outputJSON {
			data, _ := json.MarshalIndent(projects, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Projects (%d):\n\n", len(projects))
			for _, p := range projects {
				if p.ProjectTypeKey != "" {
					fmt.Printf("%-10s %-18s %s\n", p.Key, p.ProjectTypeKey, p.Name)
				} else {
					fmt.Printf("%-10s %s\n", p.Key, p.Name)
				}
			}
		}
	},
}

var jiraProjectGetCmd = &cobra.Command{
	Use:   "get <project-key>",
	Short: "Get project details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)
		project, err := client.GetProject(context.Background(), args[0])
		exitOnError(err)

		outputJSON, _ := cmd.Flags().GetBool("json")
		if outputJSON {
			data, _ := json.MarshalIndent(project, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Key:  %s\nName: %s\nID:   %s\n", project.Key, project.Name, project.ID)
		}
	},
}

var jiraWorklogCmd = &cobra.Command{
	Use:   "worklog <issue-key> <time-spent>",
	Short: "Add worklog to an issue (e.g., '2h', '30m', '1d')",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)
		err := client.AddWorklog(context.Background(), args[0], args[1])
		exitOnError(err)

		fmt.Printf("Added worklog of %s to %s\n", args[1], args[0])
	},
}

func printIssue(issue *jira.Issue) {
	fmt.Printf("Key:         %s\n", issue.Key)
	fmt.Printf("Summary:     %s\n", issue.Fields.Summary)

	if issue.Fields.Status != nil {
		fmt.Printf("Status:      %s\n", issue.Fields.Status.Name)
	}
	if issue.Fields.IssueType != nil {
		fmt.Printf("Type:        %s\n", issue.Fields.IssueType.Name)
	}
	if issue.Fields.Priority != nil {
		fmt.Printf("Priority:    %s\n", issue.Fields.Priority.Name)
	}
	if issue.Fields.Assignee != nil {
		fmt.Printf("Assignee:    %s\n", issue.Fields.Assignee.DisplayName)
	}
	if issue.Fields.Reporter != nil {
		fmt.Printf("Reporter:    %s\n", issue.Fields.Reporter.DisplayName)
	}
	if issue.Fields.Project != nil {
		fmt.Printf("Project:     %s (%s)\n", issue.Fields.Project.Name, issue.Fields.Project.Key)
	}
	if len(issue.Fields.Labels) > 0 {
		fmt.Printf("Labels:      %v\n", issue.Fields.Labels)
	}
	fmt.Printf("Created:     %s\n", issue.Fields.Created)
	fmt.Printf("Updated:     %s\n", issue.Fields.Updated)

	if issue.Fields.Description != nil {
		fmt.Printf("\nDescription:\n%v\n", issue.Fields.Description)
	}
}

func init() {
	// Issue commands
	jiraIssueGetCmd.Flags().Bool("json", false, "Output as JSON")

	jiraIssueCreateCmd.Flags().StringP("project", "p", "", "Project key (required)")
	jiraIssueCreateCmd.Flags().StringP("type", "t", "Task", "Issue type")
	jiraIssueCreateCmd.Flags().StringP("summary", "s", "", "Issue summary (required)")
	jiraIssueCreateCmd.Flags().StringP("description", "d", "", "Issue description")
	jiraIssueCreateCmd.Flags().String("priority", "", "Priority")
	jiraIssueCreateCmd.Flags().StringSlice("labels", nil, "Labels")
	jiraIssueCreateCmd.Flags().String("parent", "", "Parent issue key (for subtasks)")
	jiraIssueCreateCmd.MarkFlagRequired("project")
	jiraIssueCreateCmd.MarkFlagRequired("summary")

	jiraIssueUpdateCmd.Flags().StringP("summary", "s", "", "New summary")
	jiraIssueUpdateCmd.Flags().StringP("description", "d", "", "New description")
	jiraIssueUpdateCmd.Flags().String("priority", "", "New priority")
	jiraIssueUpdateCmd.Flags().StringSlice("labels", nil, "New labels")

	jiraIssueSearchCmd.Flags().Int("max", 50, "Maximum results")
	jiraIssueSearchCmd.Flags().Bool("json", false, "Output as JSON")

	jiraIssueCmd.AddCommand(jiraIssueGetCmd, jiraIssueCreateCmd, jiraIssueUpdateCmd, jiraIssueDeleteCmd, jiraIssueSearchCmd, jiraIssueTransitionCmd)

	// Comment commands
	jiraCommentListCmd.Flags().Bool("json", false, "Output as JSON")
	jiraCommentCmd.AddCommand(jiraCommentListCmd, jiraCommentAddCmd)

	// Project commands
	jiraProjectListCmd.Flags().Bool("json", false, "Output as JSON")
	jiraProjectListCmd.Flags().String("type", "", "Filter by project type key (e.g. product_discovery, software)")
	jiraProjectGetCmd.Flags().Bool("json", false, "Output as JSON")
	jiraProjectCmd.AddCommand(jiraProjectListCmd, jiraProjectGetCmd)

	// Add all to jira
	jiraCmd.AddCommand(jiraIssueCmd, jiraCommentCmd, jiraProjectCmd, jiraWorklogCmd)

	// Add to root
	rootCmd.AddCommand(jiraCmd)
}
