package logger

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ConfigOption func(cfg zap.Config) zap.Config

// WithCustomField добавляет постоянные поля в логи
func WithCustomField(key, value string) ConfigOption {
	return func(cfg zap.Config) zap.Config {
		if cfg.InitialFields == nil {
			cfg.InitialFields = map[string]interface{}{}
		}
		cfg.InitialFields[key] = value
		
		return cfg
	}
}

func WithStringCasting() ConfigOption {
	const encodingName = "json_string_casting"

	return func(cfg zap.Config) zap.Config {
		cfg.Encoding = encodingName
		_ = zap.RegisterEncoder(encodingName, func(config zapcore.EncoderConfig) (zapcore.Encoder, error) {
			encoder := newStringCastingEncoder(config)
			return encoder, nil
		})

		return cfg
	}
}

func WithOutputPaths(paths []string) ConfigOption {
	return func(cfg zap.Config) zap.Config {
		cfg.OutputPaths = paths
		return cfg
	}
}

type Option func(lf *loggerFormat) *loggerFormat

func WithLogRequestFunc(l func(req *http.Request)) Option {
	return func(lf *loggerFormat) *loggerFormat {
		lf.logRequestFunc = l
		return lf
	}
}

func WithLogResponseFunc(l func(context.Context, *http.Response, time.Time)) Option {
	return func(lf *loggerFormat) *loggerFormat {
		lf.logResponseFunc = l
		return lf
	}
}
