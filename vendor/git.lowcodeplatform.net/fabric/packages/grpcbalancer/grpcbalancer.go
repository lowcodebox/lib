package grpcbalancer

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	_ "github.com/Jille/grpc-multi-resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/resolver"
)

const (
	resolverDefaultScheme     = "dns"
	resolverSchemePrefixMulti = "multi:///"
)

var (
	ErrNoTarget    = errors.New("no target to dial")
	ErrSerializing = errors.New("error serializing grpc config")
	ErrEmptyTarget = errors.New("empty target")
	timeoutDefault = 1 * time.Second
)

type Client struct {
	conn      *grpc.ClientConn
	target    string
	balancing BalancingPolicy
	insecure  bool
	timeout   time.Duration
	cancel    context.CancelFunc
}

func (c *Client) Conn(ctx context.Context) (*grpc.ClientConn, error) {
	conn := c.GetConn()
	if conn != nil {
		return conn, nil
	}

	f, err := c.SetConn(ctx)
	return f, err
}

func (c *Client) GetConn() *grpc.ClientConn {
	if c.conn != nil {
		if c.conn.GetState() == connectivity.Ready {
			return c.conn
		}
		_ = c.conn.Close()
	}
	return nil
}

func (c *Client) SetConn(ctx context.Context) (*grpc.ClientConn, error) {
	if c.target == "" {
		return nil, ErrNoTarget
	}

	resolver.SetDefaultScheme(resolverDefaultScheme)
	schema := ""
	if strings.Contains(c.target, ",") {
		schema = resolverSchemePrefixMulti
	}

	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}
	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}
	conf := grpcConfig{
		Balancing: c.balancing,
	}
	confSerialized, err := json.Marshal(conf)
	if err != nil {
		return nil, ErrSerializing
	}
	opts = append(opts, grpc.WithDefaultServiceConfig(string(confSerialized)))

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.timeout)
	c.cancel = cancel
	conn, err := grpc.DialContext(ctxWithTimeout, schema+c.target, opts...)
	if err != nil {
		return nil, err
	}

	c.conn = conn
	return c.conn, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	c.cancel()
	return c.conn.Close()
}

func New(opts ...BalancerOption) (*Client, error) {
	c := &Client{}

	for _, opt := range opts {
		opt(c)
	}

	if c.timeout == 0 {
		c.timeout = timeoutDefault
	}
	if c.target == "" {
		return nil, ErrEmptyTarget
	}

	return c, nil
}
