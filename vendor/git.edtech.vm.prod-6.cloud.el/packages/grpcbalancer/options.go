package grpcbalancer

import (
	"strings"
	"time"

	"google.golang.org/grpc"
)

type BalancerOption func(*Client)

func WithUrls(target ...string) BalancerOption {
	return func(c *Client) {
		c.target = strings.Join(target, ",")
	}
}

func WithInsecure() BalancerOption {
	return func(c *Client) {
		c.insecure = true
	}
}

// WithTimeout sets connect timeout. timeout = 0 means default timeout
func WithTimeout(timeout time.Duration) BalancerOption {
	return func(c *Client) {
		c.connectTimeout = timeout
	}
}

// WithBalancingMode sets BalancingPolicy. Default is round-robin
func WithBalancingMode(bp BalancingPolicy) BalancerOption {
	return func(c *Client) {
		c.balancing = bp
	}
}

// WithForceHeathCheck sets to make health check before each request. timeout = 0 means default timeout
func WithForceHeathCheck(timeout time.Duration) BalancerOption {
	return func(c *Client) {
		c.forceHeathCheck = true
		c.healthCheckTimeout = timeout
	}
}

func WithUnaryInterceptor(interceptor grpc.UnaryClientInterceptor) BalancerOption {
	return func(c *Client) {
		c.unaryInterceptor = interceptor
	}
}

func WithChainUnaryInterceptors(interceptors ...grpc.UnaryClientInterceptor) BalancerOption {
	return func(c *Client) {
		c.chainUnaryInterceptors = interceptors
	}
}
