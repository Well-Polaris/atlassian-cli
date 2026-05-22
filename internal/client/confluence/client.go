package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/peter/atlassian-cli/internal/client"
)

// Client provides access to Confluence REST API
type Client struct {
	*client.Client
}

// New creates a new Confluence client
func New(c *client.Client) *Client {
	return &Client{Client: c}
}

// Page represents a Confluence page
type Page struct {
	ID      string     `json:"id"`
	Title   string     `json:"title"`
	Status  string     `json:"status"`
	SpaceID string     `json:"spaceId"`
	Body    *PageBody  `json:"body,omitempty"`
	Version *Version   `json:"version,omitempty"`
	Links   *PageLinks `json:"_links,omitempty"`
}

// PageBody contains the page content
type PageBody struct {
	Storage        *BodyContent `json:"storage,omitempty"`
	AtlasDocFormat *BodyContent `json:"atlas_doc_format,omitempty"`
}

// BodyContent represents page content in a specific format
type BodyContent struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

// Version represents a page version
type Version struct {
	Number    int    `json:"number"`
	Message   string `json:"message,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// PageLinks contains page links
type PageLinks struct {
	WebUI string `json:"webui"`
	Self  string `json:"self"`
}

// Space represents a Confluence space
type Space struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
}

// Comment represents a Confluence comment
type Comment struct {
	ID        string       `json:"id"`
	Status    string       `json:"status"`
	Title     string       `json:"title,omitempty"`
	Body      *CommentBody `json:"body,omitempty"`
	CreatedAt string       `json:"createdAt"`
	Version   *Version     `json:"version,omitempty"`
}

// CommentBody contains comment content
type CommentBody struct {
	Storage        *BodyContent `json:"storage,omitempty"`
	AtlasDocFormat *BodyContent `json:"atlas_doc_format,omitempty"`
}

// SearchResult represents a CQL search result
type SearchResult struct {
	Results []SearchResultItem `json:"results"`
	Start   int                `json:"start"`
	Limit   int                `json:"limit"`
	Size    int                `json:"size"`
}

// SearchResultItem represents a single search result
type SearchResultItem struct {
	Content   *Page  `json:"content,omitempty"`
	Title     string `json:"title"`
	Excerpt   string `json:"excerpt"`
	URL       string `json:"url"`
	ResultType string `json:"resultType"`
}

// PaginatedResponse wraps paginated API responses
type PaginatedResponse struct {
	Results []interface{} `json:"results"`
	Links   struct {
		Next string `json:"next,omitempty"`
	} `json:"_links"`
}

// GetPage retrieves a page by ID
func (c *Client) GetPage(ctx context.Context, pageID string, includeBody bool) (*Page, error) {
	path := fmt.Sprintf("/wiki/api/v2/pages/%s", url.PathEscape(pageID))
	if includeBody {
		path += "?body-format=storage"
	}

	var page Page
	if err := c.DoREST(ctx, "GET", path, nil, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// CreatePageInput contains fields for creating a page
type CreatePageInput struct {
	SpaceID  string
	Title    string
	Body     string
	ParentID string
	Status   string // "current" or "draft"
}

// CreatePage creates a new page
func (c *Client) CreatePage(ctx context.Context, input *CreatePageInput) (*Page, error) {
	body := map[string]interface{}{
		"spaceId": input.SpaceID,
		"title":   input.Title,
		"body": map[string]interface{}{
			"representation": "storage",
			"value":          input.Body,
		},
	}

	if input.ParentID != "" {
		body["parentId"] = input.ParentID
	}

	status := input.Status
	if status == "" {
		status = "current"
	}
	body["status"] = status

	var page Page
	if err := c.DoREST(ctx, "POST", "/wiki/api/v2/pages", body, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// UpdatePageInput contains fields for updating a page
type UpdatePageInput struct {
	Title   string
	Body    string
	Version int
	Message string
}

// UpdatePage updates an existing page
func (c *Client) UpdatePage(ctx context.Context, pageID string, input *UpdatePageInput) (*Page, error) {
	body := map[string]interface{}{
		"id":     pageID,
		"status": "current",
		"title":  input.Title,
		"body": map[string]interface{}{
			"representation": "storage",
			"value":          input.Body,
		},
		"version": map[string]interface{}{
			"number":  input.Version,
			"message": input.Message,
		},
	}

	path := fmt.Sprintf("/wiki/api/v2/pages/%s", url.PathEscape(pageID))

	var page Page
	if err := c.DoREST(ctx, "PUT", path, body, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

// DeletePage deletes a page
func (c *Client) DeletePage(ctx context.Context, pageID string) error {
	path := fmt.Sprintf("/wiki/api/v2/pages/%s", url.PathEscape(pageID))
	return c.DoREST(ctx, "DELETE", path, nil, nil)
}

// ListPagesInSpace lists pages in a space
func (c *Client) ListPagesInSpace(ctx context.Context, spaceID string, limit int) ([]*Page, error) {
	if limit <= 0 {
		limit = 25
	}

	path := fmt.Sprintf("/wiki/api/v2/spaces/%s/pages?limit=%d", url.PathEscape(spaceID), limit)

	var result struct {
		Results []*Page `json:"results"`
	}
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// ListSpaces lists all accessible spaces
func (c *Client) ListSpaces(ctx context.Context, limit int) ([]*Space, error) {
	if limit <= 0 {
		limit = 25
	}

	path := fmt.Sprintf("/wiki/api/v2/spaces?limit=%d", limit)

	var result struct {
		Results []*Space `json:"results"`
	}
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// GetSpace gets a space by ID
func (c *Client) GetSpace(ctx context.Context, spaceID string) (*Space, error) {
	path := fmt.Sprintf("/wiki/api/v2/spaces/%s", url.PathEscape(spaceID))

	var space Space
	if err := c.DoREST(ctx, "GET", path, nil, &space); err != nil {
		return nil, err
	}

	return &space, nil
}

// GetFooterComments gets footer comments for a page
func (c *Client) GetFooterComments(ctx context.Context, pageID string) ([]*Comment, error) {
	path := fmt.Sprintf("/wiki/api/v2/pages/%s/footer-comments", url.PathEscape(pageID))

	var result struct {
		Results []*Comment `json:"results"`
	}
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// GetInlineComments gets inline comments for a page
func (c *Client) GetInlineComments(ctx context.Context, pageID string) ([]*Comment, error) {
	path := fmt.Sprintf("/wiki/api/v2/pages/%s/inline-comments", url.PathEscape(pageID))

	var result struct {
		Results []*Comment `json:"results"`
	}
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return result.Results, nil
}

// AddFooterComment adds a footer comment to a page
func (c *Client) AddFooterComment(ctx context.Context, pageID string, body string) (*Comment, error) {
	reqBody := map[string]interface{}{
		"pageId": pageID,
		"body": map[string]interface{}{
			"representation": "storage",
			"value":          fmt.Sprintf("<p>%s</p>", body),
		},
	}

	var comment Comment
	if err := c.DoREST(ctx, "POST", "/wiki/api/v2/footer-comments", reqBody, &comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

// SearchCQL searches using CQL
func (c *Client) SearchCQL(ctx context.Context, cql string, limit int) (*SearchResult, error) {
	if limit <= 0 {
		limit = 25
	}

	path := fmt.Sprintf("/wiki/rest/api/search?cql=%s&limit=%d", url.QueryEscape(cql), limit)

	var result SearchResult
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Attachment represents a file attached to a page.
type Attachment struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// ListAttachments returns the attachments on a page.
func (c *Client) ListAttachments(ctx context.Context, pageID string) ([]*Attachment, error) {
	path := fmt.Sprintf("/wiki/rest/api/content/%s/child/attachment?limit=100", url.PathEscape(pageID))
	var result struct {
		Results []*Attachment `json:"results"`
	}
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result.Results, nil
}

// findAttachment returns the attachment with the given filename on a page, or
// nil if none exists.
func (c *Client) findAttachment(ctx context.Context, pageID, filename string) (*Attachment, error) {
	path := fmt.Sprintf("/wiki/rest/api/content/%s/child/attachment?filename=%s",
		url.PathEscape(pageID), url.QueryEscape(filename))
	var result struct {
		Results []*Attachment `json:"results"`
	}
	if err := c.DoREST(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	if len(result.Results) > 0 {
		return result.Results[0], nil
	}
	return nil, nil
}

// UploadAttachment uploads a file as an attachment to a page. If an attachment
// with the same filename already exists it is updated with a new version,
// otherwise a new attachment is created. The Confluence storage-format markup
// to embed it is the same in both cases: <ri:attachment ri:filename="...">.
func (c *Client) UploadAttachment(ctx context.Context, pageID, filePath string) (*Attachment, error) {
	name := filepath.Base(filePath)

	existing, err := c.findAttachment(ctx, pageID, name)
	if err != nil {
		return nil, fmt.Errorf("checking existing attachments: %w", err)
	}

	var path string
	if existing != nil {
		// Update the existing attachment's data (adds a new version).
		path = fmt.Sprintf("/wiki/rest/api/content/%s/child/attachment/%s/data",
			url.PathEscape(pageID), url.PathEscape(existing.ID))
	} else {
		path = fmt.Sprintf("/wiki/rest/api/content/%s/child/attachment", url.PathEscape(pageID))
	}

	// X-Atlassian-Token defeats XSRF checking, required for this endpoint.
	headers := map[string]string{"X-Atlassian-Token": "no-check"}
	data, err := c.DoUpload(ctx, path, "file", filePath, headers)
	if err != nil {
		return nil, err
	}

	// The create endpoint returns {"results":[...]}; the data endpoint returns
	// a single attachment object. Accept either shape.
	var wrapped struct {
		Results []*Attachment `json:"results"`
	}
	if json.Unmarshal(data, &wrapped) == nil && len(wrapped.Results) > 0 {
		return wrapped.Results[0], nil
	}
	var single Attachment
	if json.Unmarshal(data, &single) == nil && single.ID != "" {
		return &single, nil
	}
	return nil, fmt.Errorf("upload succeeded but response could not be parsed: %s", string(data))
}
