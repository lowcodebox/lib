package logbox_client

import (
	"context"
	"fmt"
	"time"

	"git.lowcodeplatform.net/packages/grpcbalancer"
)

var timeoutDefault = 1 * time.Second

type client struct {
	client *grpcbalancer.Client
}

type Client interface {
	Upsert(ctx context.Context, in upsertReq) (out upsertRes, err error)
	Search(ctx context.Context, in searchRes) (out searchReq, err error)

	NewUpsertReq() upsertReq
	NewEvent(Uid string, Level string, Type string, Name string, ConfigID string, RequestID string, ServiceID string, Msg string, Time string, Timing string, Payload string) event

	Close() error
}

func (c *client) NewUpsertReq() upsertReq {
	return upsertReq{}
}

func (c *client) Close() error {
	return c.client.Close()
}

func New(ctx context.Context, url string, reqTimeout time.Duration) (Client, error) {
	if reqTimeout == 0 {
		reqTimeout = timeoutDefault
	}
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
