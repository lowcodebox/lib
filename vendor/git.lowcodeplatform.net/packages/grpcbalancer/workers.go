package grpcbalancer

import (
	"time"
)

const (
	checkConnInterval     = 10 * time.Second
	softReconnectInterval = 10 * time.Minute
)

func (c *Client) checkConn() {
	checkConnTicker := time.NewTicker(checkConnInterval)

	for range checkConnTicker.C {
		func() {
			c.connMx.RLock()
			conn := c.conn
			c.connMx.RUnlock()

			if conn == nil {
				return
			}

			err := conn.HealthCheck()

			if err != nil {
				c.connMx.Lock()
				c.conn.ResetConnectBackoff()
				c.connMx.Unlock()
			}
		}()
	}
}

func (c *Client) softReconnect() {
	softReconnectTicker := time.NewTicker(softReconnectInterval)

	for range softReconnectTicker.C {
		func() {
			c.mx.Lock()
			defer c.mx.Unlock()

			_ = c.Close()
			_, _ = c.SetConn(c.ctx)
		}()
	}
}
