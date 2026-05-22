package goals

import (
	"context"

	"github.com/Well-Polaris/atlassian-cli/internal/client"
)

// Client provides access to Atlassian Goals GraphQL API
type Client struct {
	*client.Client
}

// New creates a new Goals client
func New(c *client.Client) *Client {
	return &Client{Client: c}
}

// Goal represents an Atlassian Goal
type Goal struct {
	ID          string        `json:"id"`
	ARI         string        `json:"ari"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	State       string        `json:"state"`
	DueDate     string        `json:"dueDate,omitempty"`
	StartDate   string        `json:"startDate,omitempty"`
	CreatedAt   string        `json:"createdAt"`
	UpdatedAt   string        `json:"updatedAt"`
	Owner       *User         `json:"owner,omitempty"`
	Metrics     []*MetricTarget `json:"metricTargets,omitempty"`
}

// User represents a user
type User struct {
	AccountID   string `json:"accountId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// MetricTarget represents a metric target on a goal
type MetricTarget struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	CurrentValue float64 `json:"currentValue"`
	TargetValue  float64 `json:"targetValue"`
	StartValue   float64 `json:"startValue"`
	Unit         string  `json:"unit,omitempty"`
}

// GoalUpdate represents an update/status on a goal
type GoalUpdate struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	State     string `json:"state"`
	CreatedAt string `json:"createdAt"`
	Author    *User  `json:"author,omitempty"`
}

// ListGoals lists all accessible goals
func (c *Client) ListGoals(ctx context.Context, limit int) ([]*Goal, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		query ListGoals($first: Int!) {
			goals(first: $first) {
				nodes {
					id
					ari
					name
					description
					state
					dueDate
					startDate
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
		Goals struct {
			Nodes []*Goal `json:"nodes"`
		} `json:"goals"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.Goals.Nodes, nil
}

// GetGoal retrieves a goal by ID
func (c *Client) GetGoal(ctx context.Context, goalID string) (*Goal, error) {
	query := `
		query GetGoal($id: ID!) {
			goal(id: $id) {
				id
				ari
				name
				description
				state
				dueDate
				startDate
				createdAt
				updatedAt
				owner {
					accountId
					name
				}
				metricTargets {
					nodes {
						id
						name
						currentValue
						targetValue
						startValue
						unit
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": goalID,
	}

	var result struct {
		Goal *Goal `json:"goal"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.Goal, nil
}

// CreateGoalInput contains fields for creating a goal
type CreateGoalInput struct {
	Name        string
	Description string
	DueDate     string // ISO 8601 format
	StartDate   string // ISO 8601 format
}

// CreateGoal creates a new goal
func (c *Client) CreateGoal(ctx context.Context, input *CreateGoalInput) (*Goal, error) {
	query := `
		mutation CreateGoal($input: CreateGoalInput!) {
			createGoal(input: $input) {
				goal {
					id
					ari
					name
					description
					state
					dueDate
					startDate
					createdAt
				}
			}
		}
	`

	goalInput := map[string]interface{}{
		"name": input.Name,
	}
	if input.Description != "" {
		goalInput["description"] = input.Description
	}
	if input.DueDate != "" {
		goalInput["dueDate"] = input.DueDate
	}
	if input.StartDate != "" {
		goalInput["startDate"] = input.StartDate
	}

	variables := map[string]interface{}{
		"input": goalInput,
	}

	var result struct {
		CreateGoal struct {
			Goal *Goal `json:"goal"`
		} `json:"createGoal"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.CreateGoal.Goal, nil
}

// UpdateGoalInput contains fields for updating a goal
type UpdateGoalInput struct {
	Name        *string
	Description *string
	DueDate     *string
	StartDate   *string
	State       *string
}

// UpdateGoal updates an existing goal
func (c *Client) UpdateGoal(ctx context.Context, goalID string, input *UpdateGoalInput) (*Goal, error) {
	query := `
		mutation UpdateGoal($id: ID!, $input: UpdateGoalInput!) {
			updateGoal(id: $id, input: $input) {
				goal {
					id
					ari
					name
					description
					state
					dueDate
					startDate
					updatedAt
				}
			}
		}
	`

	goalInput := map[string]interface{}{}
	if input.Name != nil {
		goalInput["name"] = *input.Name
	}
	if input.Description != nil {
		goalInput["description"] = *input.Description
	}
	if input.DueDate != nil {
		goalInput["dueDate"] = *input.DueDate
	}
	if input.StartDate != nil {
		goalInput["startDate"] = *input.StartDate
	}
	if input.State != nil {
		goalInput["state"] = *input.State
	}

	variables := map[string]interface{}{
		"id":    goalID,
		"input": goalInput,
	}

	var result struct {
		UpdateGoal struct {
			Goal *Goal `json:"goal"`
		} `json:"updateGoal"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.UpdateGoal.Goal, nil
}

// DeleteGoal deletes a goal
func (c *Client) DeleteGoal(ctx context.Context, goalID string) error {
	query := `
		mutation DeleteGoal($id: ID!) {
			deleteGoal(id: $id) {
				success
			}
		}
	`

	variables := map[string]interface{}{
		"id": goalID,
	}

	var result struct {
		DeleteGoal struct {
			Success bool `json:"success"`
		} `json:"deleteGoal"`
	}

	return c.DoGraphQLWithErrors(ctx, query, variables, &result)
}

// UpdateMetricValue updates a metric target's current value
func (c *Client) UpdateMetricValue(ctx context.Context, metricTargetID string, value float64) error {
	query := `
		mutation UpdateMetricTarget($id: ID!, $input: UpdateMetricTargetInput!) {
			updateMetricTarget(id: $id, input: $input) {
				metricTarget {
					id
					currentValue
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": metricTargetID,
		"input": map[string]interface{}{
			"currentValue": value,
		},
	}

	var result struct {
		UpdateMetricTarget struct {
			MetricTarget *MetricTarget `json:"metricTarget"`
		} `json:"updateMetricTarget"`
	}

	return c.DoGraphQLWithErrors(ctx, query, variables, &result)
}

// ListGoalUpdates lists updates for a goal
func (c *Client) ListGoalUpdates(ctx context.Context, goalID string, limit int) ([]*GoalUpdate, error) {
	if limit <= 0 {
		limit = 20
	}

	query := `
		query ListGoalUpdates($goalId: ID!, $first: Int!) {
			goal(id: $goalId) {
				updates(first: $first) {
					nodes {
						id
						message
						state
						createdAt
						author {
							accountId
							name
						}
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"goalId": goalID,
		"first":  limit,
	}

	var result struct {
		Goal struct {
			Updates struct {
				Nodes []*GoalUpdate `json:"nodes"`
			} `json:"updates"`
		} `json:"goal"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.Goal.Updates.Nodes, nil
}

// CreateGoalUpdate adds a status update to a goal
func (c *Client) CreateGoalUpdate(ctx context.Context, goalID string, message string, state string) (*GoalUpdate, error) {
	query := `
		mutation CreateGoalUpdate($goalId: ID!, $input: CreateGoalUpdateInput!) {
			createGoalUpdate(goalId: $goalId, input: $input) {
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
		"goalId": goalID,
		"input": map[string]interface{}{
			"message": message,
			"state":   state,
		},
	}

	var result struct {
		CreateGoalUpdate struct {
			Update *GoalUpdate `json:"update"`
		} `json:"createGoalUpdate"`
	}

	if err := c.DoGraphQLWithErrors(ctx, query, variables, &result); err != nil {
		return nil, err
	}

	return result.CreateGoalUpdate.Update, nil
}
