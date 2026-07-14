package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type Certificate struct {
	ID          int                    `json:"id"`
	CreatedOn   string                 `json:"created_on"`
	ModifiedOn  string                 `json:"modified_on"`
	OwnerUserID int                    `json:"owner_user_id"`
	Provider    string                 `json:"provider"`
	NiceName    string                 `json:"nice_name"`
	DomainNames []string               `json:"domain_names"`
	ExpiresOn   string                 `json:"expires_on"`
	Meta        map[string]interface{} `json:"meta"`
}

func (c *Client) ListCertificates(ctx context.Context) ([]Certificate, error) {
	resp, err := c.Get(ctx, "/api/nginx/certificates?expand=owner")
	if err != nil {
		return nil, err
	}
	var certs []Certificate
	if err := resp.JSON(&certs); err != nil {
		return nil, err
	}
	return certs, nil
}

func (c *Client) GetCertificate(ctx context.Context, id int) (*Certificate, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/nginx/certificates/%d", id))
	if err != nil {
		return nil, err
	}
	var cert Certificate
	if err := resp.JSON(&cert); err != nil {
		return nil, err
	}
	return &cert, nil
}

func (c *Client) CreateCertificate(ctx context.Context, body interface{}) (*Certificate, error) {
	resp, err := c.Post(ctx, "/api/nginx/certificates", body)
	if err != nil {
		return nil, err
	}
	var cert Certificate
	if err := resp.JSON(&cert); err != nil {
		return nil, err
	}
	return &cert, nil
}

func (c *Client) DeleteCertificate(ctx context.Context, id int) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/api/nginx/certificates/%d", id))
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) RenewCertificate(ctx context.Context, id int) error {
	resp, err := c.Post(ctx, fmt.Sprintf("/api/nginx/certificates/%d/renew", id), nil)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) ListDNSProviders(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.Get(ctx, "/api/nginx/certificates/dns-providers")
	if err != nil {
		return nil, err
	}
	var providers json.RawMessage
	if err := resp.JSON(&providers); err != nil {
		return nil, err
	}
	return providers, nil
}

func (c *Client) TestHTTP(ctx context.Context, body interface{}) error {
	resp, err := c.Post(ctx, "/api/nginx/certificates/test-http", body)
	if err != nil {
		return err
	}
	return resp.Error()
}
