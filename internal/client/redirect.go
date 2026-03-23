package client

import (
	"context"
	"fmt"
)

type RedirectHost struct {
	ID                int                    `json:"id"`
	CreatedOn         string                 `json:"created_on"`
	ModifiedOn        string                 `json:"modified_on"`
	OwnerUserID       int                    `json:"owner_user_id"`
	DomainNames       []string               `json:"domain_names"`
	ForwardScheme     string                 `json:"forward_scheme"`
	ForwardHTTPCode   int                    `json:"forward_http_code"`
	ForwardDomainName string                 `json:"forward_domain_name"`
	PreservePath      bool                   `json:"preserve_path"`
	CertificateID     interface{}            `json:"certificate_id"`
	SSLForced         bool                   `json:"ssl_forced"`
	HSTSEnabled       bool                   `json:"hsts_enabled"`
	HSTSSubdomains    bool                   `json:"hsts_subdomains"`
	HTTP2Support      bool                   `json:"http2_support"`
	BlockExploits     bool                   `json:"block_exploits"`
	AdvancedConfig    string                 `json:"advanced_config"`
	Enabled           bool                   `json:"enabled"`
	Meta              map[string]interface{} `json:"meta"`
}

func (c *Client) ListRedirectHosts(ctx context.Context) ([]RedirectHost, error) {
	resp, err := c.Get(ctx, "/api/nginx/redirection-hosts?expand=owner,certificate")
	if err != nil {
		return nil, err
	}
	var hosts []RedirectHost
	if err := resp.JSON(&hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func (c *Client) GetRedirectHost(ctx context.Context, id int) (*RedirectHost, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/nginx/redirection-hosts/%d", id))
	if err != nil {
		return nil, err
	}
	var host RedirectHost
	if err := resp.JSON(&host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *Client) CreateRedirectHost(ctx context.Context, body interface{}) (*RedirectHost, error) {
	resp, err := c.Post(ctx, "/api/nginx/redirection-hosts", body)
	if err != nil {
		return nil, err
	}
	var host RedirectHost
	if err := resp.JSON(&host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *Client) UpdateRedirectHost(ctx context.Context, id int, body interface{}) (*RedirectHost, error) {
	resp, err := c.Put(ctx, fmt.Sprintf("/api/nginx/redirection-hosts/%d", id), body)
	if err != nil {
		return nil, err
	}
	var host RedirectHost
	if err := resp.JSON(&host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *Client) DeleteRedirectHost(ctx context.Context, id int) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/api/nginx/redirection-hosts/%d", id))
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) EnableRedirectHost(ctx context.Context, id int) error {
	resp, err := c.Post(ctx, fmt.Sprintf("/api/nginx/redirection-hosts/%d/enable", id), nil)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) DisableRedirectHost(ctx context.Context, id int) error {
	resp, err := c.Post(ctx, fmt.Sprintf("/api/nginx/redirection-hosts/%d/disable", id), nil)
	if err != nil {
		return err
	}
	return resp.Error()
}
