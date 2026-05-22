package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/peter/atlassian-cli/internal/client/jira"
	"github.com/spf13/cobra"
)

var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "Manage issue links (works across Jira and JPD)",
	Long: "Create, list and delete links between issues. Because JPD ideas are " +
		"Jira issues, these commands link regular Jira tickets to Product Discovery " +
		"ideas (and vice versa) — something the Atlassian MCP cannot do.",
}

var linkTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List available issue link types",
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)
		types, err := client.GetIssueLinkTypes(context.Background())
		exitOnError(err)

		outputJSON, _ := cmd.Flags().GetBool("json")
		if outputJSON {
			data, _ := json.MarshalIndent(types, "", "  ")
			fmt.Println(string(data))
			return
		}

		fmt.Printf("Issue link types (%d):\n\n", len(types))
		for _, t := range types {
			fmt.Printf("%-16s  outward: %-20s  inward: %s\n", t.Name, t.Outward, t.Inward)
		}
	},
}

var linkCreateCmd = &cobra.Command{
	Use:   "create <from-key> <to-key>",
	Short: "Link two issues (e.g. a Jira ticket to a JPD idea)",
	Long: "Create a link between two issues. The link reads \"<from> <relation> <to>\" — " +
		"for type \"Blocks\", from-key blocks to-key. Either key may be a regular Jira " +
		"issue or a JPD idea. Run 'atlassian link types' to see valid --type values.",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		linkType, _ := cmd.Flags().GetString("type")
		comment, _ := cmd.Flags().GetString("comment")

		client := jira.New(apiClient)
		err := client.CreateIssueLink(context.Background(), linkType, args[0], args[1], comment)
		exitOnError(err)

		fmt.Printf("Linked %s -> %s (%s)\n", args[0], args[1], linkType)
	},
}

var linkListCmd = &cobra.Command{
	Use:   "list <issue-key>",
	Short: "List links on an issue or idea",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)
		issue, err := client.GetIssueLinks(context.Background(), args[0])
		exitOnError(err)

		outputJSON, _ := cmd.Flags().GetBool("json")
		if outputJSON {
			data, _ := json.MarshalIndent(issue.Fields.IssueLinks, "", "  ")
			fmt.Println(string(data))
			return
		}

		links := issue.Fields.IssueLinks
		fmt.Printf("Links for %s (%d):\n\n", args[0], len(links))
		for _, l := range links {
			relation, other := linkDirection(l)
			status := ""
			if other != nil && other.Fields.Status != nil {
				status = other.Fields.Status.Name
			}
			summary := ""
			if other != nil {
				summary = other.Fields.Summary
			}
			otherKey := ""
			if other != nil {
				otherKey = other.Key
			}
			fmt.Printf("[%s] %-18s %-12s [%-12s] %s\n", l.ID, relation, otherKey, status, summary)
		}
	},
}

var linkDeleteCmd = &cobra.Command{
	Use:   "delete <link-id>",
	Short: "Delete an issue link by its ID (see 'link list')",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		exitOnError(requireRESTAuth())

		client := jira.New(apiClient)
		err := client.DeleteIssueLink(context.Background(), args[0])
		exitOnError(err)

		fmt.Printf("Deleted link: %s\n", args[0])
	},
}

// linkDirection resolves how a link reads from the perspective of the queried
// issue, returning the relation phrase and the issue at the other end.
func linkDirection(l *jira.IssueLink) (string, *jira.Issue) {
	if l.OutwardIssue != nil {
		relation := "relates to"
		if l.Type != nil {
			relation = l.Type.Outward
		}
		return relation, l.OutwardIssue
	}
	relation := "relates to"
	if l.Type != nil {
		relation = l.Type.Inward
	}
	return relation, l.InwardIssue
}

func init() {
	linkTypesCmd.Flags().Bool("json", false, "Output as JSON")

	linkCreateCmd.Flags().StringP("type", "t", "Relates", "Issue link type name (see 'link types')")
	linkCreateCmd.Flags().String("comment", "", "Optional comment to add with the link")

	linkListCmd.Flags().Bool("json", false, "Output as JSON")

	linkCmd.AddCommand(linkTypesCmd, linkCreateCmd, linkListCmd, linkDeleteCmd)
	rootCmd.AddCommand(linkCmd)
}
