package projects

import (
	"context"

	"github.com/peter/atlassian-cli/internal/client"
)

// Client provides access to Atlassian Projects (Atlas) GraphQL API
type Client struct {
	*client.Client
}

// New creates a new Projects client
func New(c *client.Client) *Client {
	return &Client{Client: c}
}

// Project represents an Atlassian Project (Atlas)
type Project struct {
	ID          string   `json:"id"`
	ARI         string   `json:"ari"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	State       string   `json:"state"`
	StartDate   string   `json:"startDate,omitempty"`
	TargetDate  string   `json:"targetDate,omitempty"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
	Owner       *User    `json:"owner,omitempty"`
	Goals       []*Goal  `json:"goals,omitempty"`
}

// User represents a user
type User struct {
	AccountID   string `json:"accountId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// Goal represents a linked goal
type Goal struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ProjectUpdate represents a status update on a project
type ProjectUpdate struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
	Author    *User  `json:"author,omitempty"`
}

// ListProjects lists all accessible projects
func (c *Client) ListProjects(ctx context.Context, limit int) ([]*Project, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		query ListProjects($first: Int!) {
			projects(first: $first) {
				nodes {
					id
					ari
					name
					description
					state
					startDate
					targetDate
					createdAt
					updatedAt
					owner {
						accountId
						name
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"first": limit,
	}

	var result struct {
		Projects struct {
			Nodes []*Project `json:"nodes"`
		} `json:"projects"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.Projects.Nodes, nil
}

// GetProject retrieves a project by ID
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	query := `
		query GetProject($id: ID!) {
			project(id: $id) {
				id
				ari
				name
				description
				state
				startDate
				targetDate
				createdAt
				updatedAt
				owner {
					accountId
					name
				}
				goals {
					nodes {
						id
						name
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": projectID,
	}

	var result struct {
		Project *Project `json:"project"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.Project, nil
}

// CreateProjectInput contains fields for creating a project
type CreateProjectInput struct {
	Name        string
	Description string
	StartDate   string
	TargetDate  string
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, input *CreateProjectInput) (*Project, error) {
	query := `
		mutation CreateProject($input: CreateProjectInput!) {
			createProject(input: $input) {
				project {
					id
					ari
					name
					description
					state
					startDate
					targetDate
					createdAt
				}
			}
		}
	`

	projectInput := map[string]interface{}{
		"name": input.Name,
	}
	if input.Description != "" {
		projectInput["description"] = input.Description
	}
	if input.StartDate != "" {
		projectInput["startDate"] = input.StartDate
	}
	if input.TargetDate != "" {
		projectInput["targetDate"] = input.TargetDate
	}

	variables := map[string]interface{}{
		"input": projectInput,
	}

	var result struct {
		CreateProject struct {
			Project *Project `json:"project"`
		} `json:"createProject"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.CreateProject.Project, nil
}

// UpdateProjectInput contains fields for updating a project
type UpdateProjectInput struct {
	Name        *string
	Description *string
	StartDate   *string
	TargetDate  *string
	State       *string
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(ctx context.Context, projectID string, input *UpdateProjectInput) (*Project, error) {
	query := `
		mutation UpdateProject($id: ID!, $input: UpdateProjectInput!) {
			updateProject(id: $id, input: $input) {
				project {
					id
					ari
					name
					description
					state
					startDate
					targetDate
					updatedAt
				}
			}
		}
	`

	projectInput := map[string]interface{}{}
	if input.Name != nil {
		projectInput["name"] = *input.Name
	}
	if input.Description != nil {
		projectInput["description"] = *input.Description
	}
	if input.StartDate != nil {
		projectInput["startDate"] = *input.StartDate
	}
	if input.TargetDate != nil {
		projectInput["targetDate"] = *input.TargetDate
	}
	if input.State != nil {
		projectInput["state"] = *input.State
	}

	variables := map[string]interface{}{
		"id":    projectID,
		"input": projectInput,
	}

	var result struct {
		UpdateProject struct {
			Project *Project `json:"project"`
		} `json:"updateProject"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.UpdateProject.Project, nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	query := `
		mutation DeleteProject($id: ID!) {
			deleteProject(id: $id) {
				success
			}
		}
	`

	variables := map[string]interface{}{
		"id": projectID,
	}

	var result struct {
		DeleteProject struct {
			Success bool `json:"success"`
		} `json:"deleteProject"`
	}

	return c.DoGraphQLWithErrors(ctx, query, variables, &result)
}

// LinkGoal links a goal to a project
func (c *Client) LinkGoal(ctx context.Context, projectID, goalID string) error {
	query := `
		mutation LinkGoalToProject($projectId: ID!, $goalId: ID!) {
			linkGoalToProject(projectId: $projectId, goalId: $goalId) {
				success
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
		"goalId":    goalID,
	}

	var result struct {
		LinkGoalToProject struct {
			Success bool `json:"success"`
		} `json:"linkGoalToProject"`
	}

	return c.DoGraphQLWithErrors(ctx, query, variables, &result)
}

// UnlinkGoal unlinks a goal from a project
func (c *Client) UnlinkGoal(ctx context.Context, projectID, goalID string) error {
	query := `
		mutation UnlinkGoalFromProject($projectId: ID!, $goalId: ID!) {
			unlinkGoalFromProject(projectId: $projectId, goalId: $goalId) {
				success
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
		"goalId":    goalID,
	}

	var result struct {
		UnlinkGoalFromProject struct {
			Success bool `json:"success"`
		} `json:"unlinkGoalFromProject"`
	}

	return c.DoGraphQLWithErrors(ctx, query, variables, &result)
}

// CreateProjectUpdate adds a status update to a project
func (c *Client) CreateProjectUpdate(ctx context.Context, projectID string, message string, state string) (*ProjectUpdate, error) {
	query := `
		mutation CreateProjectUpdate($projectId: ID!, $input: CreateProjectUpdateInput!) {
			createProjectUpdate(projectId: $projectId, input: $input) {
				update {
					id
					message
					state
					createdAt
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": projectID,
		"input": map[string]interface{}{
			"message": message,
			"state":   state,
		},
	}

	var result struct {
		CreateProjectUpdate struct {
			Update *ProjectUpdate `json:"update"`
		} `json:"createProjectUpdate"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.CreateProjectUpdate.Update, nil
}
