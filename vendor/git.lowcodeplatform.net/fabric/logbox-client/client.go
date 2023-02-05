package logbox_client

import (
	"context"
	"fmt"
	"time"

	"git.lowcodeplatform.net/fabric/packages/grpcbalancer"
)

type client struct {
	client *grpcbalancer.Client
}

type Client interface {
	Upsert(ctx context.Context, in upsertReq) (out upsertRes, err error)

	NewUpsertReq() *upsertReq
	NewEvent(config, level, msg, name, srv, time, uid string) *event
}

func (c *client) NewUpsertReq() *upsertReq {
	return &upsertReq{}
}

func New(ctx context.Context, url string, reqTimeout time.Duration) (Client, error) {
	b, err := grpcbalancer.New(
		grpcbalancer.WithUrls(url),
		grpcbalancer.WithInsecure(),
		grpcbalancer.WithTimeout(reqTimeout),
	)
	if err != nil {
		fmt.Printf("failed init grpcbalancer, err: %s", err)

		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("error init connect (grpcbalancer) client")
	}
	// фикс, если нет соединения, то возвращается {}
	if fmt.Sprint(b) == "{}" {
		return nil, fmt.Errorf("error init connect (grpcbalancer) client")
	}

	return &client{
		client: b,
	}, err
}
