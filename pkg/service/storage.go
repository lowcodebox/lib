package service

import (
	"context"
	"fmt"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
)

// Storage ...
func (s *service) Storage(ctx context.Context, in model.StorageIn) (out model.StorageOut, err error) {
	defer s.monitoringTimingService("Storage", time.Now())
	defer s.monitoringError("Storage", err)

	if in.Bucket != "upload" && in.Bucket != "templates" && in.Bucket != "assets" {
		out.Body, out.MimeType, err = s.vfs.ReadFromBucket(ctx, in.File, in.Bucket) // читаем из заданного бакета (в данном случае только для templates)
	} else {
		out.Body, out.MimeType, err = s.vfs.Read(ctx, in.File)
	}

	if len(out.Body) == 0 {
		err = fmt.Errorf("file not found")
	}

	return out, err
}
