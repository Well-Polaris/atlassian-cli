package jira

import (
	"context"
	"fmt"
	"net/url"

	"github.com/peter/atlassian-cli/internal/client"
)

// Client provides access to Jira REST API
type Client struct {
	*client.Client
}

// New creates a new Jira client
func New(c *client.Client) *Client {
	return &Client{Client: c}
}

// Issue represents a Jira issue
type Issue struct {
	ID     string      `json:"id"`
	Key    string      `json:"key"`
	Self   string      `json:"self"`
	Fields IssueFields `json:"fields"`
}

// IssueFields contains the fields of an issue
type IssueFields struct {
	Summary     string      `json:"summary"`
	Description interface{} `json:"description"` // Can be string or ADF
	Status      *Status     `json:"status"`
	IssueType   *IssueType  `json:"issuetype"`
	Priority    *Priority   `json:"priority"`
	Assignee    *User       `json:"assignee"`
	Reporter    *User       `json:"reporter"`
	Created     string      `json:"created"`
	Updated     string      `json:"updated"`
	Project     *Project    `json:"project"`
	Labels      []string    `json:"labels"`
	Parent      *Issue      `json:"parent,omitempty"`
	IssueLinks  []*IssueLink `json:"issuelinks,omitempty"`
}

// Status represents an issue status
type Status struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IssueType represents an issue type
type IssueType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Priority represents an issue priority
type Priority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// User represents a Jira user
type User struct {
	AccountID   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	Email       string `json:"emailAddress"`
}

// Project represents a Jira project
type Project struct {
	ID             string `json:"id"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	ProjectTypeKey string `json:"projectTypeKey,omitempty"`
}

// IssueLinkType represents a type of issue link (e.g. "Blocks", "Relates")
type IssueLinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

// IssueLink represents a link between two issues. As seen from a given issue,
// only one of InwardIssue / OutwardIssue is populated (the other end).
type IssueLink struct {
	ID           string         `json:"id"`
	Type         *IssueLinkType `json:"type"`
	InwardIssue  *Issue         `json:"inwardIssue,omitempty"`
	OutwardIssue *Issue         `json:"outwardIssue,omitempty"`
}

// Comment represents a Jira comment
type Comment struct {
	ID      string      `json:"id"`
	Body    interface{} `json:"body"`
	Author  *User       `json:"author"`
	Created string      `json:"created"`
	Updated string      `json:"updated"`
}

// Transition represents an issue transition
type Transition struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	To   *Status `json:"to"`
}

// SearchResult represents a JQL search result. The /search/jql endpoint is
// token-paginated and does not return a total count.
type SearchResult struct {
	Issues        []*Issue `json:"issues"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
	IsLast        bool     `json:"isLast,omitempty"`
}

// GetIssue retrieves an issue by key or ID
func (c *Client) GetIssue(ctx context.Context, issueKeyOrID string) (*Issue, error) {
	path := fmt.Sprintf("/rest/api/3/issue/%s", url.PathEscape(issueKeyOrID))
	var issue Issue
	if err := c.DoREST(ctx, "GET", path, nil, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// CreateIssueInput contains fields for creating an issue
type CreateIssueInput struct {
	ProjectKey  string                 `json:"-"`
	IssueType   string                 `json:"-"`
	Summary     string                 `json:"-"`
	Description string                 `json:"-"`
	ParentKey   string                 `json:"-"`
	Priority    string                 `json:"-"`
	Labels      []string               `json:"-"`
	Assignee    string                 `json:"-"`
	Extra       map[string]interface{} `json:"-"`
}

// CreateIssue creates a new issue
func (c *Client) CreateIssue(ctx context.Context, input *CreateIssueInput) (*Issue, error) {
	fields := map[string]interface{}{
		"project":   map[string]string{"key": input.ProjectKey},
		"issuetype": map[string]string{"name": input.IssueType},
		"summary":   input.Summary,
	}

	if input.Description != "" {
		// Use ADF format for description
		fields["description"] = map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{"type": "text", "text": input.Description},
					},
				},
			},
		}
	}

	if input.ParentKey != "" {
		fields["parent"] = map[string]string{"key": input.ParentKey}
	}

	if input.Priority != "" {
		fields["priority"] = map[string]string{"name": input.Priority}
	}

	if len(input.Labels) > 0 {
		fields["labels"] = input.Labels
	}

	if input.Assignee != "" {
		fields["assignee"] = map[string]string{"accountId": input.Assignee}
	}

	for k, v := range input.Extra {
		fields[k] = v
	}

	body := map[string]interface{}{"fields": fields}

	var issue Issue
	if err := c.DoREST(ctx, "POST", "/rest/api/3/issue", body, &issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// UpdateIssueInput contains fields for updating an issue
type UpdateIssueInput struct {
	Summary     *string                `json:"-"`
	Description *string                `json:"-"`
	Priority    *string                `json:"-"`
	Labels      []string               `json:"-"`
	Assignee    *string                `json:"-"`
	Extra       map[string]interface{} `json:"-"`
}

// UpdateIssue updates an existing issue
func (c *Client) UpdateIssue(ctx context.Context, issueKeyOrID string, input *UpdateIssueInput) error {
	fields := map[string]interface{}{}

	if input.Summary != nil {
		fields["summary"] = *input.Summary
	}

	if input.Description != nil {
		fields["description"] = map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{"type": "text", "text": *input.Description},
					},
				},
			},
		}
	}

	if input.Priority != nil {
		fields["priority"] = map[string]string{"name": *input.Priority}
	}

	if input.Labels != nil {
		fields["labels"] = input.Labels
	}

	if input.Assignee != nil {
		fields["assignee"] = map[string]string{"accountId": *input.Assignee}
	}

	for k, v := range input.Extra {
		fields[k] = v
	}

	body := map[string]interface{}{"fields": fields}
	path := fmt.Sprintf("/rest/api/3/issue/%s", url.PathEscape(issueKeyOrID))

	return c.DoREST(ctx, "PUT", path, body, nil)
}

// DeleteIssue deletes an issue
func (c *Client) DeleteIssue(ctx context.Context, issueKeyOrID string) error {
	path := fmt.Sprintf("/rest/api/3/issue/%s", url.PathEscape(issueKeyOrID))
	return c.DoREST(ctx, "DELETE", path, nil, nil)
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(ctx context.Context, jql string, maxResults int) (*SearchResult, error) {
	body := map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
		"fields":     []string{"summary", "description", "status", "issuetype", "priority", "assignee", "reporter", "created", "updated", "project", "labels", "parent"},
	}

	var result SearchResult
	// /search/jql replaces the removed /search endpoint (deprecated 2024, removed May 2025).
	if err := c.DoREST(ctx, "POST", "/rest/api/3/search/jql", body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTransitions gets available transitions for an issue
func (c *Client) GetTransitions(ctx context.Context, issueKeyOrID string) ([]Transition, error) {
	path := fmt.Sprintf("/rest/api/3/issue/%s/transitions", url.PathEscape(issueKeyOrID))

	var result struct {
		Transitions []Transition `json:"transitions"`
	}
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return result.Transitions, nil
}

// TransitionIssue transitions an issue to a new status
func (c *Client) TransitionIssue(ctx context.Context, issueKeyOrID string, transitionID string) error {
	path := fmt.Sprintf("/rest/api/3/issue/%s/transitions", url.PathEscape(issueKeyOrID))
	body := map[string]interface{}{
		"transition": map[string]string{"id": transitionID},
	}

	return c.DoREST(ctx, "POST", path, body, nil)
}

// GetComments gets comments on an issue
func (c *Client) GetComments(ctx context.Context, issueKeyOrID string) ([]Comment, error) {
	path := fmt.Sprintf("/rest/api/3/issue/%s/comment", url.PathEscape(issueKeyOrID))

	var result struct {
		Comments []Comment `json:"comments"`
	}
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return result.Comments, nil
}

// AddComment adds a comment to an issue
func (c *Client) AddComment(ctx context.Context, issueKeyOrID string, body string) (*Comment, error) {
	path := fmt.Sprintf("/rest/api/3/issue/%s/comment", url.PathEscape(issueKeyOrID))
	reqBody := map[string]interface{}{
		"body": map[string]interface{}{
			"type":    "doc",
			"version": 1,
			"content": []map[string]interface{}{
				{
					"type": "paragraph",
					"content": []map[string]interface{}{
						{"type": "text", "text": body},
					},
				},
			},
		},
	}

	var comment Comment
	if err := c.DoREST(ctx, "POST", path, reqBody, &comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

// ListProjects lists all accessible projects
func (c *Client) ListProjects(ctx context.Context) ([]*Project, error) {
	var projects []*Project
	if err := c.DoREST(ctx, "GET", "/rest/api/3/project", nil, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

// GetProject gets a project by key or ID
func (c *Client) GetProject(ctx context.Context, projectKeyOrID string) (*Project, error) {
	path := fmt.Sprintf("/rest/api/3/project/%s", url.PathEscape(projectKeyOrID))
	var project Project
	if err := c.DoREST(ctx, "GET", path, nil, &project); err != nil {
		return nil, err
	}
	return &project, nil
}

// AddWorklog adds a worklog to an issue
func (c *Client) AddWorklog(ctx context.Context, issueKeyOrID string, timeSpent string) error {
	path := fmt.Sprintf("/rest/api/3/issue/%s/worklog", url.PathEscape(issueKeyOrID))
	body := map[string]interface{}{
		"timeSpent": timeSpent,
	}

	return c.DoREST(ctx, "POST", path, body, nil)
}

// SearchProjects lists projects, optionally filtered by project type key
// (e.g. "product_discovery", "software", "service_desk", "business").
// Unlike GET /project, this endpoint reliably surfaces JPD projects.
func (c *Client) SearchProjects(ctx context.Context, typeKey string) ([]*Project, error) {
	path := "/rest/api/3/project/search?maxResults=100"
	if typeKey != "" {
		path += "&typeKey=" + url.QueryEscape(typeKey)
	}

	var result struct {
		Values []*Project `json:"values"`
		IsLast bool       `json:"isLast"`
	}
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Values, nil
}

// GetIssueLinkTypes lists the issue link types available in the instance.
func (c *Client) GetIssueLinkTypes(ctx context.Context) ([]IssueLinkType, error) {
	var result struct {
		IssueLinkTypes []IssueLinkType `json:"issueLinkTypes"`
	}
	if err := c.DoREST(ctx, "GET", "/rest/api/3/issueLinkType", nil, &result); err != nil {
		return nil, err
	}
	return result.IssueLinkTypes, nil
}

// CreateIssueLink links two issues. The link reads "outwardKey <outward> inwardKey"
// — e.g. for type "Blocks", outwardKey blocks inwardKey. This works across project
// types, so it links regular Jira issues to JPD ideas and vice versa. linkType is
// the link type Name (see GetIssueLinkTypes). An optional comment is added to the
// outward issue.
func (c *Client) CreateIssueLink(ctx context.Context, linkType, outwardKey, inwardKey, comment string) error {
	body := map[string]interface{}{
		"type":         map[string]string{"name": linkType},
		"outwardIssue": map[string]string{"key": outwardKey},
		"inwardIssue":  map[string]string{"key": inwardKey},
	}

	if comment != "" {
		body["comment"] = map[string]interface{}{
			"body": map[string]interface{}{
				"type":    "doc",
				"version": 1,
				"content": []map[string]interface{}{
					{
						"type": "paragraph",
						"content": []map[string]interface{}{
							{"type": "text", "text": comment},
						},
					},
				},
			},
		}
	}

	return c.DoREST(ctx, "POST", "/rest/api/3/issueLink", body, nil)
}

// DeleteIssueLink removes an issue link by its link ID.
func (c *Client) DeleteIssueLink(ctx context.Context, linkID string) error {
	path := fmt.Sprintf("/rest/api/3/issueLink/%s", url.PathEscape(linkID))
	return c.DoREST(ctx, "DELETE", path, nil, nil)
}

// GetIssueLinks returns the links attached to an issue, along with its summary.
func (c *Client) GetIssueLinks(ctx context.Context, issueKeyOrID string) (*Issue, error) {
	path := fmt.Sprintf("/rest/api/3/issue/%s?fields=summary,issuelinks", url.PathEscape(issueKeyOrID))
	var issue Issue
	if err := c.DoREST(ctx, "GET", path, nil, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}
