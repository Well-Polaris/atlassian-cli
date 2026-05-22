package jpd

import (
	"context"

	"github.com/Well-Polaris/atlassian-cli/internal/client"
	"github.com/Well-Polaris/atlassian-cli/internal/client/jira"
)

// Client provides access to Jira Product Discovery
// JPD ideas are Jira issues, so we wrap the Jira client
type Client struct {
	*client.Client
	jira *jira.Client
}

// New creates a new JPD client
func New(c *client.Client) *Client {
	return &Client{
		Client: c,
		jira:   jira.New(c),
	}
}

// Idea represents a JPD idea (which is a Jira issue)
type Idea struct {
	ID          string   `json:"id"`
	Key         string   `json:"key"`
	Summary     string   `json:"summary"`
	Description string   `json:"description,omitempty"`
	Status      string   `json:"status"`
	IdeaType    string   `json:"ideaType,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Created     string   `json:"created"`
	Updated     string   `json:"updated"`
	Labels      []string `json:"labels,omitempty"`
	ProjectKey  string   `json:"projectKey"`
}

// Insight represents an insight attached to an idea
type Insight struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Source      string `json:"source,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

// fromJiraIssue converts a Jira issue to an Idea
func fromJiraIssue(issue *jira.Issue) *Idea {
	idea := &Idea{
		ID:      issue.ID,
		Key:     issue.Key,
		Summary: issue.Fields.Summary,
		Created: issue.Fields.Created,
		Updated: issue.Fields.Updated,
		Labels:  issue.Fields.Labels,
	}

	if issue.Fields.Status != nil {
		idea.Status = issue.Fields.Status.Name
	}
	if issue.Fields.IssueType != nil {
		idea.IdeaType = issue.Fields.IssueType.Name
	}
	if issue.Fields.Priority != nil {
		idea.Priority = issue.Fields.Priority.Name
	}
	if issue.Fields.Project != nil {
		idea.ProjectKey = issue.Fields.Project.Key
	}

	// Extract description text
	if desc, ok := issue.Fields.Description.(string); ok {
		idea.Description = desc
	}

	return idea
}

// ListIdeas lists ideas in a JPD project
func (c *Client) ListIdeas(ctx context.Context, projectKey string, maxResults int) ([]*Idea, error) {
	if maxResults <= 0 {
		maxResults = 50
	}

	// JPD uses Jira under the hood, search for issues in the project
	jql := "project = " + projectKey + " ORDER BY created DESC"
	result, err := c.jira.SearchIssues(ctx, jql, maxResults)
	if err != nil {
		return nil, err
	}

	ideas := make([]*Idea, len(result.Issues))
	for i, issue := range result.Issues {
		ideas[i] = fromJiraIssue(issue)
	}

	return ideas, nil
}

// GetIdea retrieves an idea by key
func (c *Client) GetIdea(ctx context.Context, ideaKey string) (*Idea, error) {
	issue, err := c.jira.GetIssue(ctx, ideaKey)
	if err != nil {
		return nil, err
	}

	return fromJiraIssue(issue), nil
}

// CreateIdeaInput contains fields for creating an idea
type CreateIdeaInput struct {
	ProjectKey  string
	Summary     string
	Description string
	Labels      []string
}

// CreateIdea creates a new idea in a JPD project
func (c *Client) CreateIdea(ctx context.Context, input *CreateIdeaInput) (*Idea, error) {
	// JPD ideas are created as Jira issues with type "Idea"
	jiraInput := &jira.CreateIssueInput{
		ProjectKey:  input.ProjectKey,
		IssueType:   "Idea",
		Summary:     input.Summary,
		Description: input.Description,
		Labels:      input.Labels,
	}

	issue, err := c.jira.CreateIssue(ctx, jiraInput)
	if err != nil {
		return nil, err
	}

	return fromJiraIssue(issue), nil
}

// UpdateIdeaInput contains fields for updating an idea
type UpdateIdeaInput struct {
	Summary     *string
	Description *string
	Labels      []string
}

// UpdateIdea updates an existing idea
func (c *Client) UpdateIdea(ctx context.Context, ideaKey string, input *UpdateIdeaInput) error {
	jiraInput := &jira.UpdateIssueInput{
		Summary:     input.Summary,
		Description: input.Description,
		Labels:      input.Labels,
	}

	return c.jira.UpdateIssue(ctx, ideaKey, jiraInput)
}

// DeleteIdea deletes an idea
func (c *Client) DeleteIdea(ctx context.Context, ideaKey string) error {
	return c.jira.DeleteIssue(ctx, ideaKey)
}

// GetInsights retrieves insights for an idea using GraphQL
// Note: This requires OAuth authentication
func (c *Client) GetInsights(ctx context.Context, ideaKey string) ([]*Insight, error) {
	// Insights are accessed via the Polaris GraphQL API
	query := `
		query GetInsights($issueKey: String!) {
			jira {
				issue(key: $issueKey) {
					insights {
						nodes {
							id
							description
							source
							createdAt
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"issueKey": ideaKey,
	}

	var result struct {
		Jira struct {
			Issue struct {
				Insights struct {
					Nodes []*Insight `json:"nodes"`
				} `json:"insights"`
			} `json:"issue"`
		} `json:"jira"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.Jira.Issue.Insights.Nodes, nil
}

// AddComment adds a comment to an idea
func (c *Client) AddComment(ctx context.Context, ideaKey string, body string) error {
	_, err := c.jira.AddComment(ctx, ideaKey, body)
	return err
}
