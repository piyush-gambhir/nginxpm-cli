package client

import (
	"context"
	"fmt"
)

type ProxyHost struct {
	ID                    int                    `json:"id"`
	CreatedOn             string                 `json:"created_on"`
	ModifiedOn            string                 `json:"modified_on"`
	OwnerUserID           int                    `json:"owner_user_id"`
	DomainNames           []string               `json:"domain_names"`
	ForwardHost           string                 `json:"forward_host"`
	ForwardPort           int                    `json:"forward_port"`
	ForwardScheme         string                 `json:"forward_scheme"`
	AccessListID          int                    `json:"access_list_id"`
	CertificateID         interface{}            `json:"certificate_id"` // int or "new"
	SSLForced             bool                   `json:"ssl_forced"`
	HSTSEnabled           bool                   `json:"hsts_enabled"`
	HSTSSubdomains        bool                   `json:"hsts_subdomains"`
	HTTP2Support          bool                   `json:"http2_support"`
	BlockExploits         bool                   `json:"block_exploits"`
	CachingEnabled        bool                   `json:"caching_enabled"`
	AllowWebsocketUpgrade bool                   `json:"allow_websocket_upgrade"`
	AdvancedConfig        string                 `json:"advanced_config"`
	Enabled               bool                   `json:"enabled"`
	Meta                  map[string]interface{} `json:"meta"`
	Locations             []ProxyLocation        `json:"locations"`
}

type ProxyLocation struct {
	Path           string `json:"path"`
	ForwardScheme  string `json:"forward_scheme"`
	ForwardHost    string `json:"forward_host"`
	ForwardPort    int    `json:"forward_port"`
	ForwardPath    string `json:"forward_path"`
	AdvancedConfig string `json:"advanced_config"`
}

func (c *Client) ListProxyHosts(ctx context.Context) ([]ProxyHost, error) {
	resp, err := c.Get(ctx, "/api/nginx/proxy-hosts?expand=owner,certificate,access_list")
	if err != nil {
		return nil, err
	}
	var hosts []ProxyHost
	if err := resp.JSON(&hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func (c *Client) GetProxyHost(ctx context.Context, id int) (*ProxyHost, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/nginx/proxy-hosts/%d", id))
	if err != nil {
		return nil, err
	}
	var host ProxyHost
	if err := resp.JSON(&host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *Client) CreateProxyHost(ctx context.Context, body interface{}) (*ProxyHost, error) {
	resp, err := c.Post(ctx, "/api/nginx/proxy-hosts", body)
	if err != nil {
		return nil, err
	}
	var host ProxyHost
	if err := resp.JSON(&host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *Client) UpdateProxyHost(ctx context.Context, id int, body interface{}) (*ProxyHost, error) {
	resp, err := c.Put(ctx, fmt.Sprintf("/api/nginx/proxy-hosts/%d", id), body)
	if err != nil {
		return nil, err
	}
	var host ProxyHost
	if err := resp.JSON(&host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *Client) DeleteProxyHost(ctx context.Context, id int) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/api/nginx/proxy-hosts/%d", id))
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) EnableProxyHost(ctx context.Context, id int) error {
	resp, err := c.Post(ctx, fmt.Sprintf("/api/nginx/proxy-hosts/%d/enable", id), nil)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) DisableProxyHost(ctx context.Context, id int) error {
	resp, err := c.Post(ctx, fmt.Sprintf("/api/nginx/proxy-hosts/%d/disable", id), nil)
	if err != nil {
		return err
	}
	return resp.Error()
}
