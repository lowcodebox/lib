package service

import (
	"context"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/lib"
)

// Alive ...
func (s *service) Alive(ctx context.Context) (out model.AliveOut, err error) {
	out.Cache = s.cache.Active()

	temp := s.cfg
	temp.VfsCertCA = lib.HideExceptFirstAndLast(temp.VfsCertCA)
	temp.VfsAccessKeyId = lib.HideExceptFirstAndLast(temp.VfsAccessKeyId)
	temp.VfsSecretKey = lib.HideExceptFirstAndLast(temp.VfsSecretKey)
	temp.ProjectKey = lib.HideExceptFirstAndLast(temp.ProjectKey)
	temp.VfsEndpoint = lib.HideExceptFirstAndLast(temp.VfsEndpoint)
	temp.LogboxEndpoint = lib.HideExceptFirstAndLast(temp.LogboxEndpoint)

	temp.ProjectKey = lib.HideExceptFirstAndLast(temp.ProjectKey)

	temp.LogsDir = lib.HideExceptFirstAndLast(temp.LogsDir)
	temp.Workingdir = lib.HideExceptFirstAndLast(temp.Workingdir)
	temp.ProxyPointsrc = lib.HideExceptFirstAndLast(temp.ProxyPointsrc)

	out.Config = temp

	out.Session = s.session.List()

	return
}
