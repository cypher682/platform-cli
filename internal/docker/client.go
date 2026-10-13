package docker

import (
	"context"

	"github.com/cypher682/platform-cli/internal/logs"
)

type Client struct{}

func NewClient(ctx context.Context) (*Client, error) { return &Client{}, nil }
func (c *Client) Close() error                       { return nil }
func (c *Client) Stream(ctx context.Context, service string, o logs.Options, out chan<- logs.Line) error {
	return nil
}
