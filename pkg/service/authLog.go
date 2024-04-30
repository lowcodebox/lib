package service

import (
	"context"
	"fmt"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
)

// AuthLogIn - функция авторизации (получение и создание на клиенте токена)
func (s *service) AuthLogIn(ctx context.Context, in model.ServiceAuthIn) (out model.ServiceAuthOut, err error) {
	defer s.monitoringTimingService("AuthLogIn", time.Now())
	defer s.monitoringError("AuthLogIn", err)

	status, token, err := s.iam.Auth(ctx, in.Payload, in.Ref)
	if err != nil {
		out.Error = err
		return out, fmt.Errorf("error auth from AuthLogIn. err: %s", err)
	}

	if status {
		out.XAuthToken = token
		out.Ref = in.Ref
		out.Error = nil
	}

	return out, err
}
