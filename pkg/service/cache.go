package service

import (
	"context"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"strconv"
)

func (s *service) Cache(ctx context.Context, in model.ServiceCacheIn) (out model.RestStatus, err error) {
	count, err := s.cache.Clear(in.Link)
	if err == nil {
		out.Code = "OK"
		out.Status = 200
		out.Description = "Deleted " + strconv.Itoa(count) + " objects in cache"
	}

	return out, err
}
