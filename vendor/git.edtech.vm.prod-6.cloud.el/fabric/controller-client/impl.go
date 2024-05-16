package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	ErrRespWrongType = errors.New("wrong response type. expected string")
)

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
	_, err = lib.Curl(ctx, http.MethodPost, urlc, body, nil, handlers, nil)
	return err
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
	res, err := lib.Curl(ctx, http.MethodGet, urlc+"?key="+key, "", nil, handlers, nil)
	if err != nil {
		return "", err
	}

	str, ok := res.(string)
	if !ok {
		return "", ErrRespWrongType
	}
	data := map[string]string{}
	err = json.Unmarshal([]byte(str), &data)
	if err != nil {
		return "", err
	}
	value := data["value"]
	bytes, err := base64.StdEncoding.DecodeString(value)

	return string(bytes), err
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
