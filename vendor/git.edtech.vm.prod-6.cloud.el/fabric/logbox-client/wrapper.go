package logbox_client

import (
	"context"
	"fmt"
	"time"
)

var defaultTimeout = 5 * time.Second

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

func (c *client) Upsert(ctx context.Context, in upsertReq) (_ upsertRes, _ error) {
	/*_, err = c.cb.Execute(func() (interface{}, error) {
		out, err = c.upsert(ctx, in)
		return out, err
	})*/

	go func() {
		if _, set := ctx.Deadline(); !set {
			var cf context.CancelFunc
			ctx, cf = context.WithTimeout(ctx, defaultTimeout)
			defer cf()
		}
		_, _ = c.upsert(ctx, in)
	}()

	return upsertRes{Status: true}, nil
}
