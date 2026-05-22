package compass

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/peter/atlassian-cli/internal/client"
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
