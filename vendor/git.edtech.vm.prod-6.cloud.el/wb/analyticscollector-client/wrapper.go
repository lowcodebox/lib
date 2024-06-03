package logbox_client

import (
	"context"
)

func (c *client) Search(ctx context.Context, in searchReq) (out SearchResponse, err error) {
	return c.search(ctx, in)
}

func (c *client) Set(ctx context.Context, in setReq) (out SetRes, err error) {
	return c.set(ctx, in)
}
