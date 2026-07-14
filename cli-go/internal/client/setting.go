package client

import (
	"context"
	"fmt"
)

type Setting struct {
	ID    string                 `json:"id"`
	Value interface{}            `json:"value"`
	Meta  map[string]interface{} `json:"meta"`
}

func (c *Client) ListSettings(ctx context.Context) ([]Setting, error) {
	resp, err := c.Get(ctx, "/api/settings")
	if err != nil {
		return nil, err
	}
	var settings []Setting
	if err := resp.JSON(&settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (c *Client) GetSetting(ctx context.Context, id string) (*Setting, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/settings/%s", id))
	if err != nil {
		return nil, err
	}
	var setting Setting
	if err := resp.JSON(&setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

func (c *Client) UpdateSetting(ctx context.Context, id string, body interface{}) (*Setting, error) {
	resp, err := c.Put(ctx, fmt.Sprintf("/api/settings/%s", id), body)
	if err != nil {
		return nil, err
	}
	var setting Setting
	if err := resp.JSON(&setting); err != nil {
		return nil, err
	}
	return &setting, nil
}
