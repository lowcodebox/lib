// Package lib/vfs позволяет хранить файлы на разных источниках без необходимости учитывать особенности
// каждой реализации файлового хранилища
// поддерживаются local, s3, azure (остальные активировать по-необходимости)
package lib

import (
	"context"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3_wrappers"
	"io"
	"net/http"
)

const (
	privateDirectory = "private directory"
	userUid          = "user_uid"
)

type Vfs interface {
	Item(ctx context.Context, path string) (file s3_wrappers.Item, err error)
	List(ctx context.Context, prefix string, pageSize int) (files []s3_wrappers.Item, err error)
	Read(ctx context.Context, file string, private_access bool) (data []byte, mimeType string, err error)
	ReadFromBucket(ctx context.Context, file, bucket string, private_access bool) (data []byte, mimeType string, err error)
	ReadCloser(ctx context.Context, file string, private_access bool) (reader io.ReadCloser, err error)
	ReadCloserFromBucket(ctx context.Context, file, bucket string, private_access bool) (reader io.ReadCloser, err error)
	Write(ctx context.Context, file string, data []byte) (err error)
	Delete(ctx context.Context, file string) (err error)
	Connect(ctx context.Context) (err error)
	Close() (err error)
	Proxy(trimPrefix, newPrefix string) (http.Handler, error)
}
