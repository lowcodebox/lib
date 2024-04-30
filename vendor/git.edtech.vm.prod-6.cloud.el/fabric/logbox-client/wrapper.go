package logbox_client

import (
	"context"
	"fmt"
)

func (c *client) Search(ctx context.Context, in searchRes) (out searchReq, err error) {
	_, err = c.cb.Execute(func() (interface{}, error) {
		out, err = c.search(ctx, in)
		return out, err
	})
	if err != nil {
		return out, fmt.Errorf("error request Search (primary route). check logboxCircuitBreaker. err: %s", err)
	}

	return out, err
}

func (c *client) Upsert(ctx context.Context, in upsertReq) (out upsertRes, err error) {
	_, err = c.cb.Execute(func() (interface{}, error) {
		out, err = c.upsert(ctx, in)
		return out, err
	})
	if err != nil {
		return out, fmt.Errorf("error request Upsert (primary route). check logboxCircuitBreaker. err: %s", err)
	}

	return out, err
}
