package logbox_client

import "time"

type Option func(*client)

// WithRetryInterval sets retry interval. interval == 0 means default interval. interval can't be less than timeout
func WithRetryInterval(interval time.Duration) Option {
	return func(c *client) {
		c.retryInterval = interval
	}
}

// WithRetriesCount sets retries count. count < 1 means default count
func WithRetriesCount(count int) Option {
	return func(c *client) {
		c.setRetries = count
	}
}
