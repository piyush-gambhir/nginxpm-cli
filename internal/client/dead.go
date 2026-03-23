package client

import (
	"context"
	"fmt"
)

type DeadHost struct {
	ID             int                    `json:"id"`
	CreatedOn      string                 `json:"created_on"`
	ModifiedOn     string                 `json:"modified_on"`
	OwnerUserID    int                    `json:"owner_user_id"`
	DomainNames    []string               `json:"domain_names"`
	CertificateID  interface{}            `json:"certificate_id"`
	SSLForced      bool                   `json:"ssl_forced"`
	HSTSEnabled    bool                   `json:"hsts_enabled"`
	HSTSSubdomains bool                   `json:"hsts_subdomains"`
	HTTP2Support   bool                   `json:"http2_support"`
	AdvancedConfig string                 `json:"advanced_config"`
	Enabled        bool                   `json:"enabled"`
	Meta           map[string]interface{} `json:"meta"`
}

func (c *Client) ListDeadHosts(ctx context.Context) ([]DeadHost, error) {
	resp, err := c.Get(ctx, "/api/nginx/dead-hosts?expand=owner,certificate")
	if err != nil {
		return nil, err
	}
	var hosts []DeadHost
	if err := resp.JSON(&hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func (c *Client) GetDeadHost(ctx context.Context, id int) (*DeadHost, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/nginx/dead-hosts/%d", id))
	if err != nil {
		return nil, err
	}
	var host DeadHost
	if err := resp.JSON(&host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *Client) CreateDeadHost(ctx context.Context, body interface{}) (*DeadHost, error) {
	resp, err := c.Post(ctx, "/api/nginx/dead-hosts", body)
	if err != nil {
		return nil, err
	}
	var host DeadHost
	if err := resp.JSON(&host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *Client) UpdateDeadHost(ctx context.Context, id int, body interface{}) (*DeadHost, error) {
	resp, err := c.Put(ctx, fmt.Sprintf("/api/nginx/dead-hosts/%d", id), body)
	if err != nil {
		return nil, err
	}
	var host DeadHost
	if err := resp.JSON(&host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *Client) DeleteDeadHost(ctx context.Context, id int) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/api/nginx/dead-hosts/%d", id))
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) EnableDeadHost(ctx context.Context, id int) error {
	resp, err := c.Post(ctx, fmt.Sprintf("/api/nginx/dead-hosts/%d/enable", id), nil)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) DisableDeadHost(ctx context.Context, id int) error {
	resp, err := c.Post(ctx, fmt.Sprintf("/api/nginx/dead-hosts/%d/disable", id), nil)
	if err != nil {
		return err
	}
	return resp.Error()
}
