package service

import (
	"context"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
)

// Alive ...
func (s *service) Storage(ctx context.Context, in model.StorageIn) (out model.StorageOut, err error) {
	if in.Bucket != "upload" && in.Bucket != "templates" && in.Bucket != "assets" {
		out.Body, out.MimeType, err = s.vfs.ReadFromBucket(in.File, in.Bucket) // читаем из заданного бакета (в данном случае только для templates)
	} else {
		out.Body, out.MimeType, err = s.vfs.Read(in.File)
	}

	return
}
