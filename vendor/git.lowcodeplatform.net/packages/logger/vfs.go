package logger

//
//import (
//	"bytes"
//	"context"
//	"io"
//	"sync"
//	"time"
//
//	"git.lowcodeplatform.net/fabric/lib"
//)
//
//type ConfigVfsLogger struct {
//	Kind, Endpoint, AccessKeyID, SecretKey, Region, Bucket, Comma, CACert string
//	Dir                                                                   string
//	IntervalReload                                                        time.Duration
//}
//
//// NewVfsLogger инициализация отправки логов на сервер сбора
//// ВНИМАНИЕ! крайне неэффективно
//// при добавлении лога выкачивется весь файл лога, добавляется строка и перезаписывается
//func NewVfsLogger(ctx context.Context, cfg ConfigVfsLogger) (logger Log, err error) {
//	var output io.Writer
//	m := sync.Mutex{}
//
//	vfs := lib.NewVfs(cfg.Vfs.Kind, cfg.Vfs.Endpoint, cfg.Vfs.AccessKeyID, cfg.Vfs.SecretKey, cfg.Vfs.Region, cfg.Vfs.Bucket, cfg.Vfs.Comma, cfg.Vfs.CACert)
//	err = vfs.Connect()
//	if err != nil {
//		return nil, err
//	}
//
//	sender := newVfsSender(ctx, vfs, cfg.Vfs.Dir, cfg.Srv, cfg.Uid, cfg.Vfs.IntervalReload)
//	output = sender
//
//	l := &log{
//		Output:         output,
//		Levels:         cfg.Level,
//		UID:            cfg.Uid,
//		Name:           cfg.Name,
//		Service:        cfg.Srv,
//		IntervalReload: cfg.Vfs.IntervalReload,
//		mux:            &m,
//	}
//
//	return l, nil
//}
//
//type vfsSender struct {
//	vfsStorage lib.Vfs
//	file       string
//}
//
//func (v *vfsSender) Write(p []byte) (n int, err error) {
//	dataFile, _, err := v.vfsStorage.Read(v.file)
//	concatSlices := [][]byte{
//		dataFile,
//		p,
//	}
//	resultSlice := bytes.Join(concatSlices, []byte(""))
//
//	err = v.vfsStorage.Write(v.file, resultSlice)
//	if err != nil {
//		return 0, err
//	}
//	return len(p), nil
//}
//
//func newVfsSender(ctx context.Context, vfsStorage lib.Vfs, dir, srv, uid string, intervalReload time.Duration) io.Writer {
//
//	sender := &vfsSender{
//		vfsStorage,
//		"",
//	}
//
//	//datefile := time.Now().Format("2006.01.02")
//	datefile := time.Now().Format("2006.01.02")
//	sender.file = "/" + dir + "/" + datefile + "_" + srv + "_" + uid + ".log"
//
//	// попытка обновить файл (раз в 10 минут)
//	go func() {
//		ticker := time.NewTicker(intervalReload)
//		defer ticker.Stop()
//
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-ticker.C:
//				datefile = time.Now().Format("2006.01.02")
//
//				sender.file = "/" + dir + "/" + datefile + "_" + srv + "_" + uid + ".log"
//				ticker = time.NewTicker(intervalReload)
//			}
//		}
//	}()
//
//	return sender
//}
