package grpcbalancer

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	_ "git.edtech.vm.prod-6.cloud.el/packages/grpcbalancer/multiresolver"
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
)

const defaultConnectTimeout = 1 * time.Second

type Client struct {
	conn                   GrpcConn
	ctx                    context.Context
	mx                     *sync.Mutex
	connMx                 *sync.RWMutex
	target                 string
	balancing              BalancingPolicy
	insecure               bool
	forceHeathCheck        bool
	connectTimeout         time.Duration
	healthCheckTimeout     time.Duration
	cancel                 context.CancelFunc
	unaryInterceptor       grpc.UnaryClientInterceptor
	chainUnaryInterceptors []grpc.UnaryClientInterceptor
	opts                   []grpc.DialOption
}

func (c *Client) Conn(ctx context.Context) (GrpcConn, error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	conn := c.GetConn()
	if conn != nil {
		return conn, nil
	}

	return c.SetConn(ctx)
}

func (c *Client) GetConn() GrpcConn {
	c.connMx.RLock()
	defer c.connMx.RUnlock()

	if c.conn != nil {
		if c.conn.GetState() == connectivity.Ready {
			return c.conn
		}

		_ = c.conn.Close()
	}

	return nil
}

func (c *Client) SetConn(ctx context.Context) (GrpcConn, error) {
	if c.target == "" {
		return nil, ErrNoTarget
	}

	c.connMx.Lock()
	defer c.connMx.Unlock()

	conn, cancel, err := c.connect(ctx)
	if err != nil {
		return nil, err
	}

	if c.conn == nil {
		c.conn = &connWrapper{
			mx:                 &sync.RWMutex{},
			forceHeathCheck:    c.forceHeathCheck,
			healthCheckTimeout: c.healthCheckTimeout,
		}
	}
	c.conn.SetConn(conn)
	c.cancel = cancel

	return c.conn, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}

	c.connMx.Lock()
	defer c.connMx.Unlock()

	c.cancel()
	return c.conn.Close()
}

func (c Client) connect(ctx context.Context) (*grpc.ClientConn, context.CancelFunc, error) {
	resolver.SetDefaultScheme(resolverDefaultScheme)
	schema := ""
	if strings.Contains(c.target, ",") {
		schema = resolverSchemePrefixMulti
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.connectTimeout)
	conn, err := grpc.DialContext(ctxWithTimeout, schema+c.target, c.opts...)

	return conn, cancel, err
}

func New(options ...BalancerOption) (*Client, error) {
	c := &Client{
		ctx:    context.Background(),
		mx:     &sync.Mutex{},
		connMx: &sync.RWMutex{},
	}

	for _, opt := range options {
		opt(c)
	}

	if c.connectTimeout == 0 {
		c.connectTimeout = defaultConnectTimeout
	}

	if c.target == "" {
		return nil, ErrEmptyTarget
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

	if c.unaryInterceptor != nil {
		opts = append(opts, grpc.WithUnaryInterceptor(c.unaryInterceptor))
	}

	if c.chainUnaryInterceptors != nil {
		opts = append(opts, grpc.WithChainUnaryInterceptor(c.chainUnaryInterceptors...))
	}

	c.opts = opts

	go c.checkConn()
	go c.softReconnect()

	return c, nil
}
