// Package subtpps provides access to the Token.io Sub-TPPs API.
// Sub-TPPs enable unregulated TPPs to operate under a regulated parent TPP.
package subtpps

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/iamkanishka/tokenio-client-go/internal/httpclient"
	"github.com/iamkanishka/tokenio-client-go/pkg/common"
)

// SubTpp represents a sub-TPP entity.
type SubTpp struct {
	ID              string    `json:"id"`
	Name            string    `json:"name,omitempty"`
	DisplayName     string    `json:"displayName,omitempty"`
	ParentID        string    `json:"parentId,omitempty"`
	Status          string    `json:"status,omitempty"`
	CreatedDateTime time.Time `json:"createdDateTime"`
	UpdatedDateTime time.Time `json:"updatedDateTime"`
}

// CreateSubTppRequest is the body for POST /sub-tpps.
type CreateSubTppRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
}

// CreateSubTppResponse wraps the created sub-TPP.
type CreateSubTppResponse struct {
	SubTpp SubTpp `json:"subTpp"`
}

// GetSubTppsResponse is returned by GET /sub-tpps.
type GetSubTppsResponse struct {
	SubTpps  []SubTpp         `json:"subTpps"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// GetSubTppResponse is returned by GET /sub-tpps/{id}.
type GetSubTppResponse struct {
	SubTpp SubTpp `json:"subTpp"`
}

// GetSubTppChildrenResponse is returned by GET /sub-tpps/{id}/children.
type GetSubTppChildrenResponse struct {
	SubTpps  []SubTpp         `json:"subTpps"`
	PageInfo *common.PageInfo `json:"pageInfo,omitempty"`
}

// Client exposes the Token.io Sub-TPPs API.
type Client struct {
	hc *httpclient.Client
}

// NewClient creates a SubTpps client backed by hc.
func NewClient(hc *httpclient.Client) *Client { return &Client{hc: hc} }

// CreateSubTpp creates a new sub-TPP under the calling member.
//
// POST /sub-tpps
func (c *Client) CreateSubTpp(ctx context.Context, req CreateSubTppRequest) (*SubTpp, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("subtpps: Name is required")
	}
	var out CreateSubTppResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodPost, "/sub-tpps").WithBody(req), &out); err != nil {
		return nil, err
	}
	return &out.SubTpp, nil
}

// GetSubTpps retrieves all sub-TPPs under the calling member.
//
// GET /sub-tpps
func (c *Client) GetSubTpps(ctx context.Context, limit int, offset string) (*GetSubTppsResponse, error) {
	if limit <= 0 || limit > 200 {
		return nil, fmt.Errorf("subtpps: limit must be between 1 and 200")
	}
	r := httpclient.NewRequest(http.MethodGet, "/sub-tpps").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetSubTppsResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSubTpp retrieves a single sub-TPP by ID.
//
// GET /sub-tpps/{id}
func (c *Client) GetSubTpp(ctx context.Context, id string) (*SubTpp, error) {
	if id == "" {
		return nil, fmt.Errorf("subtpps: id is required")
	}
	var out GetSubTppResponse
	if err := c.hc.Do(ctx, httpclient.NewRequest(http.MethodGet, "/sub-tpps/"+id), &out); err != nil {
		return nil, err
	}
	return &out.SubTpp, nil
}

// DeleteSubTpp deletes a sub-TPP.
//
// DELETE /sub-tpps/{id}
func (c *Client) DeleteSubTpp(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("subtpps: id is required")
	}
	return c.hc.Do(ctx, httpclient.NewRequest(http.MethodDelete, "/sub-tpps/"+id), nil)
}

// GetSubTppChildren retrieves child sub-TPPs of a given sub-TPP.
//
// GET /sub-tpps/{id}/children
func (c *Client) GetSubTppChildren(ctx context.Context, id string, limit int, offset string) (*GetSubTppChildrenResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("subtpps: id is required")
	}
	r := httpclient.NewRequest(http.MethodGet, "/sub-tpps/"+id+"/children").
		WithQuery("limit", strconv.Itoa(limit)).
		WithQuery("offset", offset)
	var out GetSubTppChildrenResponse
	if err := c.hc.Do(ctx, r, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
