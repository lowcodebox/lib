package grpcbalancer

import (
	"strings"
	"time"
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

func WithTimeout(time time.Duration) BalancerOption {
	return func(c *Client) {
		c.timeout = time
	}
}

// Default mode — round robin
func WithBalancingMode(bp BalancingPolicy) BalancerOption {
	return func(c *Client) {
		c.balancing = bp
	}
}
