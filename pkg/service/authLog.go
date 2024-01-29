package service

import (
	"context"
	"fmt"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
)

// AuthLogIn - функция авторизации (получение и создание на клиенте токена)
func (s *service) AuthLogIn(ctx context.Context, in model.ServiceAuthIn) (out model.ServiceAuthOut, err error) {
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
