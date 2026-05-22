package compass

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Well-Polaris/atlassian-cli/internal/client"
)

// Client provides access to the Compass GraphQL API. Compass is GraphQL-only
// and goes through the api.atlassian.com/graphql gateway with OAuth.
type Client struct {
	*client.Client
}

// New creates a new Compass client.
func New(c *client.Client) *Client {
	return &Client{Client: c}
}

// Component is a Compass component (a service, library, application, etc.).
type Component struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	State       string `json:"state"`
	TypeID      string `json:"typeId"`
	URL         string `json:"url"`
	OwnerID     string `json:"ownerId"`
}

// Scorecard is a Compass scorecard.
type Scorecard struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Description string `json:"description"`
}

// SearchComponents searches a site's Compass catalog. An empty query text
// returns all components (up to limit).
func (c *Client) SearchComponents(ctx context.Context, cloudID, query string, limit int) ([]*Component, error) {
	const q = `query SearchComponents($cloudId: String!, $query: CompassSearchComponentQuery) {
  compass {
    searchComponents(cloudId: $cloudId, query: $query) {
      __typename
      ... on CompassSearchComponentConnection {
        totalCount
        nodes { component { id name slug description state typeId url ownerId } }
      }
      ... on QueryError { message }
    }
  }
}`
	vars := map[string]interface{}{
		"cloudId": cloudID,
		"query":   map[string]interface{}{"query": query, "first": limit},
	}

	var resp struct {
		Compass struct {
			SearchComponents struct {
				TypeName string `json:"__typename"`
				Message  string `json:"message"`
				Nodes    []struct {
					Component *Component `json:"component"`
				} `json:"nodes"`
			} `json:"searchComponents"`
		} `json:"compass"`
	}
	if err := c.DoGraphQLWithErrors(ctx, q, vars, &resp); err != nil {
		return nil, err
	}

	sc := resp.Compass.SearchComponents
	if sc.TypeName == "QueryError" {
		return nil, fmt.Errorf("compass: %s", sc.Message)
	}
	out := make([]*Component, 0, len(sc.Nodes))
	for _, n := range sc.Nodes {
		if n.Component != nil {
			out = append(out, n.Component)
		}
	}
	return out, nil
}

// GetComponent retrieves a single component by its Compass ID (an ARI).
func (c *Client) GetComponent(ctx context.Context, id string) (*Component, error) {
	const q = `query GetComponent($id: ID!) {
  compass {
    component(id: $id) {
      __typename
      ... on CompassComponent { id name slug description state typeId url ownerId }
      ... on QueryError { message }
    }
  }
}`
	var resp struct {
		Compass struct {
			Component struct {
				TypeName string `json:"__typename"`
				Message  string `json:"message"`
				Component
			} `json:"component"`
		} `json:"compass"`
	}
	if err := c.DoGraphQLWithErrors(ctx, q, map[string]interface{}{"id": id}, &resp); err != nil {
		return nil, err
	}
	if resp.Compass.Component.TypeName == "QueryError" {
		return nil, fmt.Errorf("compass: %s", resp.Compass.Component.Message)
	}
	comp := resp.Compass.Component.Component
	return &comp, nil
}

// ListScorecards lists the scorecards in a site's Compass catalog.
func (c *Client) ListScorecards(ctx context.Context, cloudID string, limit int) ([]*Scorecard, error) {
	const q = `query ListScorecards($cloudId: ID!, $query: CompassScorecardsQuery) {
  compass {
    scorecards(cloudId: $cloudId, query: $query) {
      __typename
      ... on CompassScorecardConnection {
        totalCount
        nodes { id name state description }
      }
      ... on QueryError { message }
    }
  }
}`
	vars := map[string]interface{}{
		"cloudId": cloudID,
		"query":   map[string]interface{}{"first": limit},
	}

	var resp struct {
		Compass struct {
			Scorecards struct {
				TypeName string       `json:"__typename"`
				Message  string       `json:"message"`
				Nodes    []*Scorecard `json:"nodes"`
			} `json:"scorecards"`
		} `json:"compass"`
	}
	if err := c.DoGraphQLWithErrors(ctx, q, vars, &resp); err != nil {
		return nil, err
	}
	if resp.Compass.Scorecards.TypeName == "QueryError" {
		return nil, fmt.Errorf("compass: %s", resp.Compass.Scorecards.Message)
	}
	return resp.Compass.Scorecards.Nodes, nil
}

// mutationError is a business-logic error returned inside a Compass mutation
// payload (distinct from GraphQL transport errors).
type mutationError struct {
	Message string `json:"message"`
}

func joinErrors(errs []mutationError) string {
	if len(errs) == 0 {
		return "unknown error"
	}
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Message
	}
	return strings.Join(msgs, "; ")
}

// componentMutationPayload is the shared shape of the create/update component
// mutation results.
type componentMutationPayload struct {
	Success          bool            `json:"success"`
	Errors           []mutationError `json:"errors"`
	ComponentDetails *Component      `json:"componentDetails"`
}

// CreateComponentInput holds the editable fields for a new component. Name is
// required; empty fields are omitted from the request.
type CreateComponentInput struct {
	Name        string
	TypeID      string
	Description string
	Slug        string
	OwnerID     string
	State       string
}

// CreateComponent creates a new component in a site's Compass catalog.
func (c *Client) CreateComponent(ctx context.Context, cloudID string, in CreateComponentInput) (*Component, error) {
	input := map[string]interface{}{"name": in.Name}
	for k, v := range map[string]string{
		"typeId":      in.TypeID,
		"description": in.Description,
		"slug":        in.Slug,
		"ownerId":     in.OwnerID,
		"state":       in.State,
	} {
		if v != "" {
			input[k] = v
		}
	}

	const q = `mutation CreateComponent($cloudId: ID!, $input: CreateCompassComponentInput!) {
  compass {
    createComponent(cloudId: $cloudId, input: $input) {
      success
      errors { message }
      componentDetails { id name slug description state typeId url ownerId }
    }
  }
}`
	var resp struct {
		Compass struct {
			CreateComponent componentMutationPayload `json:"createComponent"`
		} `json:"compass"`
	}
	vars := map[string]interface{}{"cloudId": cloudID, "input": input}
	if err := c.DoGraphQLWithErrors(ctx, q, vars, &resp); err != nil {
		return nil, err
	}
	r := resp.Compass.CreateComponent
	if !r.Success {
		return nil, fmt.Errorf("compass rejected the create: %s", joinErrors(r.Errors))
	}
	return r.ComponentDetails, nil
}

// UpdateComponent updates an existing component. changes holds only the fields
// to modify (keyed by their GraphQL input names: name, slug, description,
// state, ownerId).
func (c *Client) UpdateComponent(ctx context.Context, id string, changes map[string]interface{}) (*Component, error) {
	input := map[string]interface{}{"id": id}
	for k, v := range changes {
		input[k] = v
	}

	const q = `mutation UpdateComponent($input: UpdateCompassComponentInput!) {
  compass {
    updateComponent(input: $input) {
      success
      errors { message }
      componentDetails { id name slug description state typeId url ownerId }
    }
  }
}`
	var resp struct {
		Compass struct {
			UpdateComponent componentMutationPayload `json:"updateComponent"`
		} `json:"compass"`
	}
	if err := c.DoGraphQLWithErrors(ctx, q, map[string]interface{}{"input": input}, &resp); err != nil {
		return nil, err
	}
	r := resp.Compass.UpdateComponent
	if !r.Success {
		return nil, fmt.Errorf("compass rejected the update: %s", joinErrors(r.Errors))
	}
	return r.ComponentDetails, nil
}

// RawQuery runs an arbitrary GraphQL operation against the Atlassian gateway
// and returns the raw response. The gateway requires a *named* operation
// (e.g. "query Name { ... }" or "mutation Name { ... }"). This is the escape
// hatch for Compass operations the typed methods do not cover, including all
// mutations (creating components, sending metric values, etc.).
func (c *Client) RawQuery(ctx context.Context, query string, vars map[string]interface{}) (json.RawMessage, error) {
	var raw json.RawMessage
	if err := c.DoGraphQL(ctx, query, vars, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
