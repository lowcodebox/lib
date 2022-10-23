package service

import (
	"context"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
)

// Alive ...
func (s *service) Alive(ctx context.Context) (out model.AliveOut, err error) {
	out.Config = s.cfg
	out.Cache = s.cache.Active()

	out.Session = s.session.List()

	return
}
