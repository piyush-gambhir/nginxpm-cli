package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type AccessList struct {
	ID          int                    `json:"id"`
	CreatedOn   string                 `json:"created_on"`
	ModifiedOn  string                 `json:"modified_on"`
	OwnerUserID int                    `json:"owner_user_id"`
	Name        string                 `json:"name"`
	SatisfyAny  bool                   `json:"satisfy_any"`
	PassAuth    bool                   `json:"pass_auth"`
	Items       json.RawMessage        `json:"items"`
	Clients     json.RawMessage        `json:"clients"`
	Meta        map[string]interface{} `json:"meta"`
}

func (c *Client) ListAccessLists(ctx context.Context) ([]AccessList, error) {
	resp, err := c.Get(ctx, "/api/nginx/access-lists?expand=owner,items,clients")
	if err != nil {
		return nil, err
	}
	var lists []AccessList
	if err := resp.JSON(&lists); err != nil {
		return nil, err
	}
	return lists, nil
}

func (c *Client) GetAccessList(ctx context.Context, id int) (*AccessList, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/nginx/access-lists/%d?expand=owner,items,clients", id))
	if err != nil {
		return nil, err
	}
	var list AccessList
	if err := resp.JSON(&list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) CreateAccessList(ctx context.Context, body interface{}) (*AccessList, error) {
	resp, err := c.Post(ctx, "/api/nginx/access-lists", body)
	if err != nil {
		return nil, err
	}
	var list AccessList
	if err := resp.JSON(&list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) UpdateAccessList(ctx context.Context, id int, body interface{}) (*AccessList, error) {
	resp, err := c.Put(ctx, fmt.Sprintf("/api/nginx/access-lists/%d", id), body)
	if err != nil {
		return nil, err
	}
	var list AccessList
	if err := resp.JSON(&list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) DeleteAccessList(ctx context.Context, id int) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/api/nginx/access-lists/%d", id))
	if err != nil {
		return err
	}
	return resp.Error()
}
