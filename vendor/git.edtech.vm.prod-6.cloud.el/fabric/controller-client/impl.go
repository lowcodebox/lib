package controller

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"

	"git.edtech.vm.prod-6.cloud.el/packages/logger"
)

var (
	ErrSecretNotFound = errors.New("secret not found")
)

type secret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type secretsOut struct {
	Secrets []secret `json:"secrets"`
	Error   error    `json:"error"`
}

func (c *controller) upsertSecret(ctx context.Context, key, value string) error {
	handlers := map[string]string{}
	err := c.generateKey(handlers)
	if err != nil {
		return err
	}

	if c.observeLog {
		defer c.observeLogger(ctx, time.Now(), "Upsert", err, key, value)
	}

	urlc, err := url.JoinPath(c.url, "secret")
	if err != nil {
		return err
	}

	body := secretsInJSON(key, value)
	var resp secretsOut
	_, err = lib.Curl(ctx, http.MethodPost, urlc, body, &resp, handlers, nil)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return resp.Error
	}
	return nil
}

func (c *controller) getSecret(ctx context.Context, key string) (string, error) {
	handlers := map[string]string{}
	err := c.generateKey(handlers)
	if err != nil {
		return "", err
	}

	if c.observeLog {
		defer c.observeLogger(ctx, time.Now(), "Get", err, key)
	}

	urlc, err := url.JoinPath(c.url, "secret")
	if err != nil {
		return "", err
	}

	key = url.QueryEscape(key)
	var resp secretsOut
	_, err = lib.Curl(ctx, http.MethodGet, urlc+"?key="+key, "", &resp, handlers, nil)
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", resp.Error
	}

	if len(resp.Secrets) == 0 {
		return "", ErrSecretNotFound
	}

	bytes, err := base64.StdEncoding.DecodeString(resp.Secrets[0].Value)

	return string(bytes), err
}

func (c *controller) listSecrets(ctx context.Context) (res map[string]string, err error) {
	handlers := map[string]string{}
	err = c.generateKey(handlers)
	if err != nil {
		return nil, err
	}

	if c.observeLog {
		defer c.observeLogger(ctx, time.Now(), "Get", err)
	}

	urlc, err := url.JoinPath(c.url, "secret", "list")
	if err != nil {
		return nil, err
	}

	var resp secretsOut
	_, err = lib.Curl(ctx, http.MethodGet, urlc, "", &resp, handlers, nil)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	res = make(map[string]string)
	for _, secret := range resp.Secrets {
		bytes, err := base64.StdEncoding.DecodeString(secret.Value)
		if err != nil {
			return nil, err
		}
		res[secret.Key] = string(bytes)
	}
	return
}

func (c *controller) generateKey(handlers map[string]string) (err error) {
	token, err := lib.GenXServiceKey(c.domain, []byte(c.projectKey), tokenInterval)
	if err != nil {
		return fmt.Errorf("error GenXServiceKey. err: %w", err)
	}
	handlers[headerServiceKey] = token
	return nil
}

func secretsInJSON(key, value string) string {
	value = base64.StdEncoding.EncodeToString([]byte(value))
	key = url.QueryEscape(key)
	return fmt.Sprintf(`{"key":"%s","value":"%s"}`, key, value)
}

func (c *controller) observeLogger(ctx context.Context, start time.Time, method string, err error, arguments ...interface{}) {
	logger.Info(ctx, "timing controller query",
		zap.String("method", method),
		zap.Float64("timing", time.Since(start).Seconds()),
		//zap.String("arguments", fmt.Sprint(arguments)),
		zap.Error(err),
	)
}
