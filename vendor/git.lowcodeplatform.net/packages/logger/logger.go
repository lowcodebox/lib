package logger

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	defaultLogger     *Engine
	level             zap.AtomicLevel
	onceLevelObserver sync.Once
)

const sep = string(os.PathSeparator)

// LogLine структура строк лог-файла. нужна для анмаршалинга
type LogLine struct {
	Uid       string      `json:"uid"`
	Level     string      `json:"level"`
	Name      string      `json:"logger"`
	Type      string      `json:"service-type"`
	Time      string      `json:"ts"`
	Timing    string      `json:"timing"`
	ConfigID  string      `json:"config-id"`
	RequestID string      `json:"request-id"`
	ServiceID string      `json:"service-id"`
	Msg       interface{} `json:"msg"`
}

type Engine struct {
	*zap.Logger
}

//goland:noinspection GoUnusedExportedFunction
func SetupDefaultLogger(namespace string, options ...ConfigOption) {
	logger := initLogger(options...)
	defaultLogger = New(logger.Named(namespace))
}

func SetupDefaultKafkaLogger(namespace string, cfg KafkaConfig) error {
	if len(cfg.Addr) == 0 {
		return errors.New("kafka address must be specified")
	}

	if err := cfg.createTopic(); err != nil {
		return errors.Wrapf(err, "cannot create topic: %s", cfg.Topic)
	}

	errorLogger := initLogger(WithStringCasting())

	ws := &writerSyncer{
		kwr:         cfg.writer(errorLogger),
		topic:       cfg.Topic,
		errorLogger: errorLogger,
	}

	enc := newStringCastingEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, ws, zap.NewAtomicLevelAt(zap.InfoLevel))

	errOut, _, err := zap.Open("stderr")
	if err != nil {
		return err
	}

	opts := []zap.Option{zap.ErrorOutput(errOut), zap.AddCaller()}

	logger := zap.New(core, opts...)
	defaultLogger = New(logger.Named(namespace))

	return nil
}

// SetupDefaultLogboxLogger инициируем логирование в сервис Logbox
func SetupDefaultLogboxLogger(namespace string, cfg LogboxConfig, options map[string]string) error {
	if len(cfg.Endpoint) == 0 {
		return errors.New("logbox address must be specified")
	}

	// инициализировать лог и его ротацию
	ws := &logboxSender{
		requestTimeout: cfg.RequestTimeout,
		logboxClient:   cfg.client(context.Background()),
	}

	enc := newStringCastingEncoder(zap.NewProductionEncoderConfig())
	core := zapcore.NewCore(enc, ws, zap.NewAtomicLevelAt(zap.InfoLevel))

	errOut, _, err := zap.Open("stderr")
	if err != nil {
		return err
	}

	opts := []zap.Option{zap.ErrorOutput(errOut), zap.AddCaller()}
	for k, v := range options {
		opts = append(opts, zap.Fields(zap.String(k, v)))
	}

	logger := zap.New(core, opts...)
	defaultLogger = New(logger.Named(namespace))

	return nil
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
