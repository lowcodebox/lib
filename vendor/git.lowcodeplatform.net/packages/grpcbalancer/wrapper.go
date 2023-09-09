package grpcbalancer

import (
	"context"
	"errors"
	"net"
	"regexp"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/status"
)

const (
	defaultHealthCheckTimeout = 1 * time.Second
	defaultRequestTimeout     = 30 * time.Second
)

var reMethod = regexp.MustCompile(`^/([^/]+)/.+$`)

type GrpcConn interface {
	WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool
	GetState() connectivity.State
	Connect()
	ResetConnectBackoff()
	Close() error
	NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error)
	Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error
	Target() string
	GetMethodConfig(method string) grpc.MethodConfig
	HealthCheck(opts ...grpc.CallOption) (err error)
	SetConn(conn *grpc.ClientConn)
}

type connWrapper struct {
	conn               *grpc.ClientConn
	mx                 *sync.RWMutex
	forceHeathCheck    bool
	healthCheckTimeout time.Duration
	healthCheckURI     string
}

func (c *connWrapper) SetConn(conn *grpc.ClientConn) {
	c.mx.Lock()
	c.conn = conn
	c.mx.Unlock()
}

func (c *connWrapper) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	var (
		cancel context.CancelFunc
		ctxTTL = ctx
	)
	if _, set := ctx.Deadline(); !set {
		ctxTTL, cancel = context.WithTimeout(ctx, defaultRequestTimeout)
		defer cancel()
	}

	c.mx.RLock()
	healthCheckURI := c.healthCheckURI
	forceHeathCheck := c.forceHeathCheck
	c.mx.RUnlock()

	if healthCheckURI == "" {
		c.mx.Lock()
		c.healthCheckURI = makeHeathCheckURI(method)
		c.mx.Unlock()
	}

	if forceHeathCheck {
		if err := c.HealthCheck(opts...); err != nil {
			return err
		}
	}

	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.conn.Invoke(ctxTTL, method, args, reply, opts...)
}

func (c *connWrapper) HealthCheck(opts ...grpc.CallOption) (err error) {
	c.mx.RLock()
	healthCheckURI := c.healthCheckURI
	healthCheckTimeout := c.healthCheckTimeout
	c.mx.RUnlock()

	if healthCheckURI == "" {
		return nil
	}

	if healthCheckTimeout == 0 {
		c.mx.Lock()
		c.healthCheckTimeout = defaultHealthCheckTimeout
		c.mx.Unlock()
	}

	checkCtx, cancel := context.WithTimeout(context.Background(), c.healthCheckTimeout)
	defer cancel()

	var checkResp interface{}
	c.mx.RLock()
	err = c.conn.Invoke(checkCtx, c.healthCheckURI, nil, checkResp, opts...)
	c.mx.RUnlock()

	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, net.ErrClosed):
		_ = c.Close()
		return status.Error(codes.Unavailable, "service health check not available")
	case err != nil:
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.Unimplemented {
				return nil
			}
		}
		return err
	}

	return nil
}

func makeHeathCheckURI(method string) string {
	return reMethod.ReplaceAllString(method, "/$1/HeathCheck")
}

func (c *connWrapper) WaitForStateChange(ctx context.Context, sourceState connectivity.State) bool {
	return c.conn.WaitForStateChange(ctx, sourceState)
}

func (c *connWrapper) GetState() connectivity.State {
	return c.conn.GetState()
}

func (c *connWrapper) Connect() {
	c.conn.Connect()
}

func (c *connWrapper) ResetConnectBackoff() {
	c.conn.ResetConnectBackoff()
}

func (c *connWrapper) Close() error {
	c.mx.Lock()
	defer c.mx.Unlock()

	return c.conn.Close()
}

func (c *connWrapper) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return c.conn.NewStream(ctx, desc, method, opts...)
}

func (c *connWrapper) Target() string {
	return c.conn.Target()
}

func (c *connWrapper) GetMethodConfig(method string) grpc.MethodConfig {
	return c.conn.GetMethodConfig(method)
}
