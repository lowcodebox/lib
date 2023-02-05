package lib

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"
)

type ConfigVfsLogger struct {
	Kind, Endpoint, AccessKeyID, SecretKey, Region, Bucket, Comma string
	Dir                                                           string
	IntervalReload                                                time.Duration
}

// NewVfsLogger инициализация отправки логов на сервер сбора
// ВНИМАНИЕ! крайне неэффективно
// при добавлении лога выкачивется весь файл лога, добавляется строка и перезаписывается
func NewVfsLogger(ctx context.Context, cfg ConfigLogger) (logger Log, err error) {
	var output io.Writer
	m := sync.Mutex{}

	vfs := NewVfs(cfg.Vfs.Kind, cfg.Vfs.Endpoint, cfg.Vfs.AccessKeyID, cfg.Vfs.SecretKey, cfg.Vfs.Region, cfg.Vfs.Bucket, cfg.Vfs.Comma)
	err = vfs.Connect()
	if err != nil {
		return nil, err
	}

	datefile := time.Now().Format("2006.01.02")
	logName := cfg.Vfs.Dir + "/" + datefile + "_" + cfg.Srv + "_" + cfg.Uid + ".log"

	sender := newVfsSender(vfs, logName)
	output = sender

	l := &log{
		Output:         output,
		Levels:         cfg.Level,
		UID:            cfg.Uid,
		Name:           cfg.Name,
		Service:        cfg.Srv,
		IntervalReload: cfg.Vfs.IntervalReload,
		mux:            &m,
	}

	return l, nil
}

type vfsSender struct {
	vfsStorage Vfs
	file       string
}

func (v *vfsSender) Write(p []byte) (n int, err error) {
	dataFile, _, err := v.vfsStorage.Read(v.file)
	concatSlices := [][]byte{
		dataFile,
		p,
	}
	resultSlice := bytes.Join(concatSlices, []byte(""))

	err = v.vfsStorage.Write(v.file, resultSlice)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func newVfsSender(vfsStorage Vfs, file string) io.Writer {
	return &vfsSender{
		vfsStorage,
		file,
	}
}
