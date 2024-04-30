package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	logboxclient "git.edtech.vm.prod-6.cloud.el/fabric/logbox-client"
)

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
	UserID    string      `json:"user-id"`
	ServiceID string      `json:"service-id"`
	Msg       interface{} `json:"msg"`
}

type LogboxConfig struct {
	Endpoint, AccessKeyID, SecretKey string
	RequestTimeout                   time.Duration
	CbMaxRequests                    uint32
	CbTimeout, CbInterval            time.Duration
	ProjectKey                       string
}

func (l *LogboxConfig) client(ctx context.Context) (client logboxclient.Client, err error) {
	client, err = logboxclient.New(ctx, l.Endpoint, l.RequestTimeout, l.CbMaxRequests, l.CbTimeout, l.CbInterval, l.ProjectKey)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type logboxSender struct {
	requestTimeout time.Duration
	logboxClient   logboxclient.Client
}

func (v *logboxSender) Write(p []byte) (n int, err error) {
	reqTimeout, cancel := context.WithTimeout(context.Background(), v.requestTimeout)
	defer cancel()

	newReq := v.logboxClient.NewUpsertReq()
	recordsBytes := bytes.Split(p, []byte("\n"))
	for _, value := range recordsBytes {
		if string(value) == "" {
			continue
		}

		l := LogLine{}
		err = json.Unmarshal(value, &l)
		if err != nil {
			return 0, fmt.Errorf("error unmarshal to logline. err: %s, value: %s", err, string(value))
		}
		newReq.AddEvent(v.logboxClient.NewEvent(
			l.Uid,
			l.Level,
			l.Type,
			l.Name,
			l.ConfigID,
			l.RequestID,
			l.UserID,
			l.ServiceID,
			l.Msg.(string),
			l.Time,
			l.Timing,
			string(value),
		))
	}

	_, err = v.logboxClient.Upsert(reqTimeout, newReq)

	return len(p), err
}

func (v *logboxSender) Sync() error {
	return v.logboxClient.Close()
}

// SetupDefaultLogboxLogger инициируем логирование в сервис Logbox
func SetupDefaultLogboxLogger(namespace string, cfg LogboxConfig, options map[string]string) (err error) {
	if len(cfg.Endpoint) == 0 {
		return errors.New("logbox address must be specified")
	}

	cl, err := cfg.client(context.Background())
	if err != nil {
		return err
	}

	// инициализировать лог и его ротацию
	ws := &logboxSender{
		requestTimeout: cfg.RequestTimeout,
		logboxClient:   cl,
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
