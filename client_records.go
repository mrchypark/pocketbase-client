package pocketbase

import (
	"context"
)

var _ RecordServiceAPI = (*Client)(nil)

func (c *Client) GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error) {
	return c.Records.GetList(ctx, collection, opts)
}

func (c *Client) GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error) {
	return c.Records.GetOne(ctx, collection, recordID, opts)
}

func (c *Client) Create(ctx context.Context, collection string, body any) (*Record, error) {
	return c.Records.Create(ctx, collection, body)
}

func (c *Client) CreateWithOptions(ctx context.Context, collection string, body any, opts *WriteOptions) (*Record, error) {
	return c.Records.CreateWithOptions(ctx, collection, body, opts)
}

func (c *Client) Update(ctx context.Context, collection, recordID string, body any) (*Record, error) {
	return c.Records.Update(ctx, collection, recordID, body)
}

func (c *Client) UpdateWithOptions(ctx context.Context, collection, recordID string, body any, opts *WriteOptions) (*Record, error) {
	return c.Records.UpdateWithOptions(ctx, collection, recordID, body, opts)
}

func (c *Client) Delete(ctx context.Context, collection, recordID string) error {
	return c.Records.Delete(ctx, collection, recordID)
}

func (c *Client) NewCreateRequest(collection string, body map[string]any) (*BatchRequest, error) {
	return c.Records.NewCreateRequest(collection, body)
}

func (c *Client) NewUpdateRequest(collection, recordID string, body map[string]any) (*BatchRequest, error) {
	return c.Records.NewUpdateRequest(collection, recordID, body)
}

func (c *Client) NewDeleteRequest(collection, recordID string) (*BatchRequest, error) {
	return c.Records.NewDeleteRequest(collection, recordID)
}

func (c *Client) NewUpsertRequest(collection string, body map[string]any) (*BatchRequest, error) {
	return c.Records.NewUpsertRequest(collection, body)
}
