package search

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Well-Polaris/atlassian-cli/internal/client"
)

// Client provides access to Atlassian unified search (Rovo)
type Client struct {
	*client.Client
}

// New creates a new Search client
func New(c *client.Client) *Client {
	return &Client{Client: c}
}

// SearchResult represents a search result
type SearchResult struct {
	ID          string          `json:"id"`
	ARI         string          `json:"ari"`
	Title       string          `json:"title"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url"`
	Type        string          `json:"type"`
	Container   *Container      `json:"container,omitempty"`
	LastUpdated string          `json:"lastUpdated,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

// Container represents the parent container (space, project, etc.)
type Container struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// SearchResponse represents the search API response
type SearchResponse struct {
	Results    []*SearchResult `json:"results"`
	TotalCount int             `json:"totalCount"`
	HasMore    bool            `json:"hasMore"`
}

// Search performs a unified search across Jira and Confluence
func (c *Client) Search(ctx context.Context, query string, limit int) (*SearchResponse, error) {
	if limit <= 0 {
		limit = 25
	}

	// Use the Rovo search API
	gqlQuery := `
		query Search($query: String!, $first: Int!) {
			search(query: $query, first: $first) {
				nodes {
					... on JiraIssue {
						id
						ari
						key
						summary
						description
						url
						project {
							id
							name
						}
						updated
					}
					... on ConfluencePage {
						id
						ari
						title
						excerpt
						url
						space {
							id
							name
						}
						lastModified
					}
				}
				totalCount
				pageInfo {
					hasNextPage
				}
			}
		}
	`

	variables := map[string]interface{}{
		"query": query,
		"first": limit,
	}

	var result struct {
		Search struct {
			Nodes      []json.RawMessage `json:"nodes"`
			TotalCount int               `json:"totalCount"`
			PageInfo   struct {
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"search"`
	}

	if err := c.DoGraphQLWithErrors(ctx, gqlQuery, variables, &result); err != nil {
		return nil, err
	}

	response := &SearchResponse{
		TotalCount: result.Search.TotalCount,
		HasMore:    result.Search.PageInfo.HasNextPage,
	}

	// Parse the union type results
	for _, raw := range result.Search.Nodes {
		sr, err := parseSearchResult(raw)
		if err != nil {
			continue // Skip unparseable results
		}
		response.Results = append(response.Results, sr)
	}

	return response, nil
}

func parseSearchResult(raw json.RawMessage) (*SearchResult, error) {
	// Try to determine the type
	var typeCheck struct {
		Key     string `json:"key"`     // Jira issues have key
		Title   string `json:"title"`   // Confluence pages have title
		Summary string `json:"summary"` // Jira issues have summary
	}

	if err := json.Unmarshal(raw, &typeCheck); err != nil {
		return nil, err
	}

	if typeCheck.Key != "" {
		// It's a Jira issue
		var issue struct {
			ID          string `json:"id"`
			ARI         string `json:"ari"`
			Key         string `json:"key"`
			Summary     string `json:"summary"`
			Description string `json:"description"`
			URL         string `json:"url"`
			Updated     string `json:"updated"`
			Project     *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"project"`
		}
		if err := json.Unmarshal(raw, &issue); err != nil {
			return nil, err
		}

		sr := &SearchResult{
			ID:          issue.ID,
			ARI:         issue.ARI,
			Title:       fmt.Sprintf("[%s] %s", issue.Key, issue.Summary),
			Description: issue.Description,
			URL:         issue.URL,
			Type:        "jira-issue",
			LastUpdated: issue.Updated,
			Raw:         raw,
		}
		if issue.Project != nil {
			sr.Container = &Container{
				ID:   issue.Project.ID,
				Name: issue.Project.Name,
				Type: "project",
			}
		}
		return sr, nil
	}

	if typeCheck.Title != "" {
		// It's a Confluence page
		var page struct {
			ID           string `json:"id"`
			ARI          string `json:"ari"`
			Title        string `json:"title"`
			Excerpt      string `json:"excerpt"`
			URL          string `json:"url"`
			LastModified string `json:"lastModified"`
			Space        *struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"space"`
		}
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, err
		}

		sr := &SearchResult{
			ID:          page.ID,
			ARI:         page.ARI,
			Title:       page.Title,
			Description: page.Excerpt,
			URL:         page.URL,
			Type:        "confluence-page",
			LastUpdated: page.LastModified,
			Raw:         raw,
		}
		if page.Space != nil {
			sr.Container = &Container{
				ID:   page.Space.ID,
				Name: page.Space.Name,
				Type: "space",
			}
		}
		return sr, nil
	}

	return nil, fmt.Errorf("unknown result type")
}

// FetchByARI retrieves a resource by its Atlassian Resource Identifier
func (c *Client) FetchByARI(ctx context.Context, ari string) (json.RawMessage, error) {
	query := `
		query Fetch($ari: ID!) {
			node(id: $ari) {
				... on JiraIssue {
					id
					ari
					key
					summary
					description
					status { name }
					priority { name }
					assignee { displayName }
					reporter { displayName }
					created
					updated
					url
				}
				... on ConfluencePage {
					id
					ari
					title
					body { storage { value } }
					status
					version { number }
					lastModified
					url
				}
				... on AtlassianGoal {
					id
					ari
					name
					description
					state
					dueDate
				}
				... on AtlassianProject {
					id
					ari
					name
					description
					state
					targetDate
				}
			}
		}
	`

	variables := map[string]interface{}{
		"ari": ari,
	}

	var result struct {
		Node json.RawMessage `json:"node"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.Node, nil
}
