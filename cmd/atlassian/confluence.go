package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/peter/atlassian-cli/internal/client/confluence"
	"github.com/spf13/cobra"
)

var confluenceCmd = &cobra.Command{
	Use:   "confluence",
	Short: "Confluence commands",
	Long:  "Commands for interacting with Confluence REST API",
}

var confluencePageCmd = &cobra.Command{
	Use:   "page",
	Short: "Page commands",
}

var confluencePageGetCmd = &cobra.Command{
	Use:   "get <page-id>",
	Short: "Get a page by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		includeBody, _ := cmd.Flags().GetBool("body")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := confluence.New(apiClient)
		page, err := client.GetPage(context.Background(), args[0], includeBody)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(page, "", "  ")
			fmt.Println(string(data))
		} else {
			printPage(page)
		}
	},
}

var confluencePageListCmd = &cobra.Command{
	Use:   "list --space <space-id>",
	Short: "List pages in a space",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		spaceID, _ := cmd.Flags().GetString("space")
		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := confluence.New(apiClient)
		pages, err := client.ListPagesInSpace(context.Background(), spaceID, limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(pages, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Pages in space %s (%d):\n\n", spaceID, len(pages))
			for _, p := range pages {
				fmt.Printf("%-12s %s\n", p.ID, p.Title)
			}
		}
	},
}

var confluencePageCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new page",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		spaceID, _ := cmd.Flags().GetString("space")
		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		parentID, _ := cmd.Flags().GetString("parent")
		status, _ := cmd.Flags().GetString("status")

		input := &confluence.CreatePageInput{
			SpaceID:  spaceID,
			Title:    title,
			Body:     body,
			ParentID: parentID,
			Status:   status,
		}

		client := confluence.New(apiClient)
		page, err := client.CreatePage(context.Background(), input)
		exitOnError(err)

		fmt.Printf("Created page: %s (ID: %s)\n", page.Title, page.ID)
		if page.Links != nil && page.Links.WebUI != "" {
			fmt.Printf("URL: %s%s\n", cfg.SiteURL, page.Links.WebUI)
		}
	},
}

var confluencePageUpdateCmd = &cobra.Command{
	Use:   "update <page-id>",
	Short: "Update an existing page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		version, _ := cmd.Flags().GetInt("version")
		message, _ := cmd.Flags().GetString("message")

		// If version not specified, get current version
		client := confluence.New(apiClient)
		if version == 0 {
			page, err := client.GetPage(context.Background(), args[0], false)
			exitOnError(err)
			if page.Version != nil {
				version = page.Version.Number + 1
			} else {
				version = 1
			}
			if title == "" {
				title = page.Title
			}
		}

		input := &confluence.UpdatePageInput{
			Title:   title,
			Body:    body,
			Version: version,
			Message: message,
		}

		page, err := client.UpdatePage(context.Background(), args[0], input)
		exitOnError(err)

		fmt.Printf("Updated page: %s (ID: %s, v%d)\n", page.Title, page.ID, page.Version.Number)
	},
}

var confluencePageDeleteCmd = &cobra.Command{
	Use:   "delete <page-id>",
	Short: "Delete a page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := confluence.New(apiClient)
		err := client.DeletePage(context.Background(), args[0])
		exitOnError(err)

		fmt.Printf("Deleted page: %s\n", args[0])
	},
}

var confluenceSpaceCmd = &cobra.Command{
	Use:   "space",
	Short: "Space commands",
}

var confluenceSpaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List spaces",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := confluence.New(apiClient)
		spaces, err := client.ListSpaces(context.Background(), limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(spaces, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Spaces (%d):\n\n", len(spaces))
			for _, s := range spaces {
				fmt.Printf("%-12s %-10s %s\n", s.ID, s.Key, s.Name)
			}
		}
	},
}

var confluenceSpaceGetCmd = &cobra.Command{
	Use:   "get <space-id>",
	Short: "Get space details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		outputJSON, _ := cmd.Flags().GetBool("json")

		client := confluence.New(apiClient)
		space, err := client.GetSpace(context.Background(), args[0])
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(space, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("ID:     %s\n", space.ID)
			fmt.Printf("Key:    %s\n", space.Key)
			fmt.Printf("Name:   %s\n", space.Name)
			fmt.Printf("Type:   %s\n", space.Type)
			fmt.Printf("Status: %s\n", space.Status)
		}
	},
}

var confluenceCommentCmd = &cobra.Command{
	Use:   "comment",
	Short: "Comment commands",
}

var confluenceCommentListCmd = &cobra.Command{
	Use:   "list <page-id>",
	Short: "List comments on a page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		inline, _ := cmd.Flags().GetBool("inline")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := confluence.New(apiClient)

		var comments []*confluence.Comment
		var err error
		if inline {
			comments, err = client.GetInlineComments(context.Background(), args[0])
		} else {
			comments, err = client.GetFooterComments(context.Background(), args[0])
		}
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(comments, "", "  ")
			fmt.Println(string(data))
		} else {
			commentType := "Footer"
			if inline {
				commentType = "Inline"
			}
			fmt.Printf("%s comments on %s (%d):\n\n", commentType, args[0], len(comments))
			for _, c := range comments {
				fmt.Printf("[%s] ID: %s\n", c.CreatedAt, c.ID)
				if c.Body != nil && c.Body.Storage != nil {
					fmt.Printf("%s\n\n", c.Body.Storage.Value)
				}
			}
		}
	},
}

var confluenceCommentAddCmd = &cobra.Command{
	Use:   "add <page-id> <comment>",
	Short: "Add a footer comment to a page",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := confluence.New(apiClient)
		comment, err := client.AddFooterComment(context.Background(), args[0], args[1])
		exitOnError(err)

		fmt.Printf("Added comment (ID: %s) to page %s\n", comment.ID, args[0])
	},
}

var confluenceSearchCmd = &cobra.Command{
	Use:   "search <cql>",
	Short: "Search using CQL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		limit, _ := cmd.Flags().GetInt("limit")
		outputJSON, _ := cmd.Flags().GetBool("json")

		client := confluence.New(apiClient)
		result, err := client.SearchCQL(context.Background(), args[0], limit)
		exitOnError(err)

		if outputJSON {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Search results (%d):\n\n", result.Size)
			for _, r := range result.Results {
				fmt.Printf("%-12s %s\n", r.ResultType, r.Title)
				if r.Excerpt != "" {
					fmt.Printf("             %s\n", r.Excerpt)
				}
				fmt.Println()
			}
		}
	},
}

func printPage(page *confluence.Page) {
	fmt.Printf("ID:      %s\n", page.ID)
	fmt.Printf("Title:   %s\n", page.Title)
	fmt.Printf("Status:  %s\n", page.Status)
	fmt.Printf("Space:   %s\n", page.SpaceID)
	if page.Version != nil {
		fmt.Printf("Version: %d\n", page.Version.Number)
	}
	if page.Links != nil && page.Links.WebUI != "" {
		fmt.Printf("URL:     %s%s\n", cfg.SiteURL, page.Links.WebUI)
	}
	if page.Body != nil && page.Body.Storage != nil {
		fmt.Printf("\nContent:\n%s\n", page.Body.Storage.Value)
	}
}

func init() {
	// Page commands
	confluencePageGetCmd.Flags().Bool("body", false, "Include page body")
	confluencePageGetCmd.Flags().Bool("json", false, "Output as JSON")

	confluencePageListCmd.Flags().String("space", "", "Space ID (required)")
	confluencePageListCmd.Flags().Int("limit", 25, "Maximum results")
	confluencePageListCmd.Flags().Bool("json", false, "Output as JSON")
	confluencePageListCmd.MarkFlagRequired("space")

	confluencePageCreateCmd.Flags().String("space", "", "Space ID (required)")
	confluencePageCreateCmd.Flags().String("title", "", "Page title (required)")
	confluencePageCreateCmd.Flags().String("body", "", "Page body (HTML/storage format)")
	confluencePageCreateCmd.Flags().String("parent", "", "Parent page ID")
	confluencePageCreateCmd.Flags().String("status", "current", "Page status (current or draft)")
	confluencePageCreateCmd.MarkFlagRequired("space")
	confluencePageCreateCmd.MarkFlagRequired("title")

	confluencePageUpdateCmd.Flags().String("title", "", "New title")
	confluencePageUpdateCmd.Flags().String("body", "", "New body (HTML/storage format)")
	confluencePageUpdateCmd.Flags().Int("version", 0, "Version number (auto-increment if not specified)")
	confluencePageUpdateCmd.Flags().String("message", "", "Version message")

	confluencePageCmd.AddCommand(confluencePageGetCmd, confluencePageListCmd, confluencePageCreateCmd, confluencePageUpdateCmd, confluencePageDeleteCmd)

	// Space commands
	confluenceSpaceListCmd.Flags().Int("limit", 25, "Maximum results")
	confluenceSpaceListCmd.Flags().Bool("json", false, "Output as JSON")
	confluenceSpaceGetCmd.Flags().Bool("json", false, "Output as JSON")
	confluenceSpaceCmd.AddCommand(confluenceSpaceListCmd, confluenceSpaceGetCmd)

	// Comment commands
	confluenceCommentListCmd.Flags().Bool("inline", false, "Show inline comments instead of footer comments")
	confluenceCommentListCmd.Flags().Bool("json", false, "Output as JSON")
	confluenceCommentCmd.AddCommand(confluenceCommentListCmd, confluenceCommentAddCmd)

	// Search command
	confluenceSearchCmd.Flags().Int("limit", 25, "Maximum results")
	confluenceSearchCmd.Flags().Bool("json", false, "Output as JSON")

	// Add all to confluence
	confluenceCmd.AddCommand(confluencePageCmd, confluenceSpaceCmd, confluenceCommentCmd, confluenceSearchCmd)

	// Add to root
	rootCmd.AddCommand(confluenceCmd)
}
