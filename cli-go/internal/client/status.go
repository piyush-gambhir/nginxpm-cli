package client

import (
	"context"
	"net/http"
)

type Status struct {
	Status  string `json:"status"`
	Version struct {
		Major    int `json:"major"`
		Minor    int `json:"minor"`
		Revision int `json:"revision"`
	} `json:"version"`
	Setup bool `json:"setup"`
}

func (c *Client) GetStatus(ctx context.Context) (*Status, error) {
	resp, err := c.doNoAuth(ctx, http.MethodGet, "/api/", nil)
	if err != nil {
		return nil, err
	}
	var status Status
	if err := resp.JSON(&status); err != nil {
		return nil, err
	}
	return &status, nil
}
