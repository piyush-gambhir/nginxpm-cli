package client

import (
	"context"
	"fmt"
)

type AuditEntry struct {
	ID         int                    `json:"id"`
	CreatedOn  string                 `json:"created_on"`
	ModifiedOn string                 `json:"modified_on"`
	UserID     int                    `json:"user_id"`
	ObjectType string                 `json:"object_type"`
	ObjectID   int                    `json:"object_id"`
	Action     string                 `json:"action"`
	Meta       map[string]interface{} `json:"meta"`
}

func (c *Client) ListAuditLog(ctx context.Context) ([]AuditEntry, error) {
	resp, err := c.Get(ctx, "/api/audit-log")
	if err != nil {
		return nil, err
	}
	var entries []AuditEntry
	if err := resp.JSON(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (c *Client) GetAuditEntry(ctx context.Context, id int) (*AuditEntry, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/audit-log/%d", id))
	if err != nil {
		return nil, err
	}
	var entry AuditEntry
	if err := resp.JSON(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}
