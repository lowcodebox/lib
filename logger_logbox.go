package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	logboxclient "git.lowcodeplatform.net/fabric/logbox-client"
)

type ConfigLogboxLogger struct {
	Endpoint, AccessKeyID, SecretKey string
	Level, Uid, Name, Srv, Config    string
	RequestTimeout                   time.Duration
}

// NewLogboxLogger инициализация логер, которых отправляет логи на сервер сбора (logbox)
func NewLogboxLogger(ctx context.Context, cfg ConfigLogboxLogger) (logger Log, err error) {
	var output io.Writer
	m := sync.Mutex{}

	client, err := logboxclient.New(ctx, cfg.Endpoint, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("error create connection whith logbox-server. err: %s", err)
	}

	sender := newLogboxSender(client, cfg.RequestTimeout)
	output = sender

	l := &log{
		Output:  output,
		Levels:  cfg.Level,
		UID:     cfg.Uid,
		Name:    cfg.Name,
		Service: cfg.Srv,
		mux:     &m,
	}

	return l, err
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
		l := LogLine{}
		err = json.Unmarshal(value, &l)
		if err != nil {
			return 0, fmt.Errorf("error unmarshal to logline. err: %s", err)
		}
		newReq.AddEvent(*v.logboxClient.NewEvent(l.Config, l.Level, l.Msg.(string), l.Name, l.Srv, l.Time, l.Uid))
	}

	_, err = v.logboxClient.Upsert(reqTimeout, *newReq)

	return len(p), err
}

func newLogboxSender(logboxClient logboxclient.Client, requestTimeout time.Duration) io.Writer {
	return &logboxSender{
		requestTimeout,
		logboxClient,
	}
}
