package controller

import (
	"context"
	"net/url"
	"strings"
	"time"
)

const headerRequestId = "X-Request-Id"
const headerServiceKey = "X-Service-Key"
const tokenInterval = 1 * time.Minute

type controller struct {
	url        string
	observeLog bool
	domain     string
	projectKey string
}

type Controller interface {
	UpsertSecret(ctx context.Context, key string, value string) error
	GetSecret(ctx context.Context, key string) (string, error)
}

func (c *controller) UpsertSecret(ctx context.Context, key string, value string) error {
	return c.upsertSecret(ctx, key, value)
}

func (c *controller) GetSecret(ctx context.Context, key string) (string, error) {
	return c.getSecret(ctx, key)
}

func New(urlStr string, observeLog bool, projectKey string) Controller {

	u, _ := url.Parse(urlStr)
	splitUrl := strings.Split(u.Path, "/")

	domain := []string{"controller", "proxy"}
	if len(splitUrl) >= 3 {
		domain = splitUrl[1:3]
	}

	return &controller{
		url:        urlStr,
		observeLog: observeLog,
		domain:     strings.Join(domain, "/"),
		projectKey: projectKey,
	}
}
