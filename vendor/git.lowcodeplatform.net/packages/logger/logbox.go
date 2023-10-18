package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	logboxclient "git.lowcodeplatform.net/fabric/logbox-client"
)

type LogboxConfig struct {
	Endpoint, AccessKeyID, SecretKey string
	RequestTimeout                   time.Duration
}

func (l *LogboxConfig) client(ctx context.Context) (client logboxclient.Client) {
	var err error
	client, err = logboxclient.New(ctx, l.Endpoint, l.RequestTimeout)
	if err != nil {
		return nil
	}
	return client
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
