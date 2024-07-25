package logger

import (
	"context"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	defaultLogger     *Engine
	level             zap.AtomicLevel
	onceLevelObserver sync.Once
)

type Engine struct {
	*zap.Logger
}

func Logger(ctx context.Context) *Engine {
	if defaultLogger == nil {
		panic("logger has not been initialized, call SetupDefaultLogger() before use")
	}

	return defaultLogger.WithContext(ctx)
}

func (l *Engine) SetLevel(lvl zapcore.Level) {
	level.SetLevel(lvl)
}

func New(logger *zap.Logger) *Engine {
	return &Engine{
		Logger: logger,
	}
}

func (l *Engine) WithContext(ctx context.Context) *Engine {
	if ctx == nil {
		return l
	}

	logger := l.Logger

	mtx.RLock()
	for field := range logKeys {
		fieldName := string(field)
		fieldNameParts := strings.Split(fieldName, ".")
		fieldName = fieldNameParts[len(fieldNameParts)-1]

		value := ctx.Value(field)
		if value == nil {
			continue
		}

		if valueStr, ok := value.(string); ok && valueStr != "" {
			logger = logger.With(zap.String(fieldName, valueStr))
		}
	}
	mtx.RUnlock()

	// оборачиваем полями из fieldStorage
	logger = withStorageFields(ctx, logger)

	return &Engine{Logger: logger}
}

func (l *Engine) AddHook(f func(zapcore.Entry)) {
	l.Logger = l.Logger.WithOptions(zap.Hooks(func(entry zapcore.Entry) error {
		f(entry)
		return nil
	}))
}

func initLogger(options ...ConfigOption) *zap.Logger {
	level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config := zap.NewProductionConfig()
	config.Level = level
	config.Sampling = &zap.SamplingConfig{
		Initial:    1000,
		Thereafter: 10,
	}

	for _, opt := range options {
		config = opt(config)
	}

	logger, _ := config.Build()

	return logger
}
