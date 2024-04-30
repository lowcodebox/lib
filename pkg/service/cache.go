package service

import (
	"context"
	"strconv"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
)

func (s *service) Cache(ctx context.Context, in model.ServiceCacheIn) (out model.RestStatus, err error) {
	defer s.monitoringTimingService("Cache", time.Now())
	defer s.monitoringError("Cache", err)

	count, err := s.cache.Clear(in.Link)
	if err == nil {
		out.Code = "OK"
		out.Status = 200
		out.Description = "Deleted " + strconv.Itoa(count) + " objects in cache"
	}

	return out, err
}
