package client

import (
	"context"
	"fmt"
)

type Stream struct {
	ID             int                    `json:"id"`
	CreatedOn      string                 `json:"created_on"`
	ModifiedOn     string                 `json:"modified_on"`
	OwnerUserID    int                    `json:"owner_user_id"`
	IncomingPort   int                    `json:"incoming_port"`
	ForwardingHost string                 `json:"forwarding_host"`
	ForwardingPort int                    `json:"forwarding_port"`
	TCPForwarding  bool                   `json:"tcp_forwarding"`
	UDPForwarding  bool                   `json:"udp_forwarding"`
	CertificateID  interface{}            `json:"certificate_id"`
	Enabled        bool                   `json:"enabled"`
	Meta           map[string]interface{} `json:"meta"`
}

func (c *Client) ListStreams(ctx context.Context) ([]Stream, error) {
	resp, err := c.Get(ctx, "/api/nginx/streams?expand=owner")
	if err != nil {
		return nil, err
	}
	var streams []Stream
	if err := resp.JSON(&streams); err != nil {
		return nil, err
	}
	return streams, nil
}

func (c *Client) GetStream(ctx context.Context, id int) (*Stream, error) {
	resp, err := c.Get(ctx, fmt.Sprintf("/api/nginx/streams/%d", id))
	if err != nil {
		return nil, err
	}
	var stream Stream
	if err := resp.JSON(&stream); err != nil {
		return nil, err
	}
	return &stream, nil
}

func (c *Client) CreateStream(ctx context.Context, body interface{}) (*Stream, error) {
	resp, err := c.Post(ctx, "/api/nginx/streams", body)
	if err != nil {
		return nil, err
	}
	var stream Stream
	if err := resp.JSON(&stream); err != nil {
		return nil, err
	}
	return &stream, nil
}

func (c *Client) UpdateStream(ctx context.Context, id int, body interface{}) (*Stream, error) {
	resp, err := c.Put(ctx, fmt.Sprintf("/api/nginx/streams/%d", id), body)
	if err != nil {
		return nil, err
	}
	var stream Stream
	if err := resp.JSON(&stream); err != nil {
		return nil, err
	}
	return &stream, nil
}

func (c *Client) DeleteStream(ctx context.Context, id int) error {
	resp, err := c.Delete(ctx, fmt.Sprintf("/api/nginx/streams/%d", id))
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) EnableStream(ctx context.Context, id int) error {
	resp, err := c.Post(ctx, fmt.Sprintf("/api/nginx/streams/%d/enable", id), nil)
	if err != nil {
		return err
	}
	return resp.Error()
}

func (c *Client) DisableStream(ctx context.Context, id int) error {
	resp, err := c.Post(ctx, fmt.Sprintf("/api/nginx/streams/%d/disable", id), nil)
	if err != nil {
		return err
	}
	return resp.Error()
}
