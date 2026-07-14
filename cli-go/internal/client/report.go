package client

import (
	"context"
	"encoding/json"
)

func (c *Client) GetHostReport(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/api/reports/hosts")
	if err != nil {
		return nil, err
	}
	var report json.RawMessage
	if err := resp.JSON(&report); err != nil {
		return nil, err
	}
	return report, nil
}
