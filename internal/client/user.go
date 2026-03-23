package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type User struct {
	ID          int              `json:"id"`
	CreatedOn   string           `json:"created_on"`
	ModifiedOn  string           `json:"modified_on"`
	Name        string           `json:"name"`
	Nickname    string           `json:"nickname"`
	Email       string           `json:"email"`
	Avatar      string           `json:"avatar"`
	IsDisabled  bool             `json:"is_disabled"`
	Roles       []string         `json:"roles"`
	Permissions json.RawMessage  `json:"permissions"`
}

func (c *Client) ListUsers(ctx context.Context) ([]User, error) {
	resp, err := c.Get(ctx, "/api/users?expand=permissions")
	if err != nil {
		return nil, err
	}
	var users []User
	if err := resp.JSON(&users); err != nil {
		return nil, err
	}
	return users, nil
}

func (c *Client) GetUser(ctx context.Context, id string) (*User, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/users/%s", id))
	if err != nil {
		return nil, err
	}
	var user User
	if err := resp.JSON(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) CreateUser(ctx context.Context, body interface{}) (*User, error) {
	resp, err := c.Post(ctx, "/api/users", body)
	if err != nil {
		return nil, err
	}
	var user User
	if err := resp.JSON(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) UpdateUser(ctx context.Context, id int, body interface{}) (*User, error) {
	resp, err := c.Put(ctx, fmt.Sprintf("/api/users/%d", id), body)
	if err != nil {
		return nil, err
	}
	var user User
	if err := resp.JSON(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *Client) DeleteUser(ctx context.Context, id int) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/api/users/%d", id))
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) SetUserPermissions(ctx context.Context, id int, body interface{}) error {
	resp, err := c.Put(ctx, fmt.Sprintf("/api/users/%d/permissions", id), body)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) ChangePassword(ctx context.Context, id int, body interface{}) error {
	resp, err := c.Put(ctx, fmt.Sprintf("/api/users/%d/auth", id), body)
	if err != nil {
		return err
	}
	return resp.Error()
}
