package service

import (
	"context"
	"fmt"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
)

// AuthChangeRole - функция обновления токена с новой ролью
// для обновления роли:
// 1. делаем запрос на IAM с передачей роли, которую ходим использовать.
// IAM проверяет возможность у данного токена обработать данную роль и обновляет данные сессии
// в ответе возвращает валидный, но завершенный токен, чтобы инициировать процесс обновления токена
// на стороне AuthProcessor приложения, атакже данные новой роли в хранилище сессии, поскольку токен будет завершенны
// 2. обновляем токен на клиенте через редирект на страницу с новой кукой,
// в которой полученный от IAM валидный, но завершенный токен
func (s *service) AuthChangeRole(ctx context.Context, in model.ServiceAuthChangeIn) (out model.ServiceAuthChangeOut, err error) {
	defer s.timingService("AuthChangeRole", time.Now())
	defer s.errorMetric("AuthChangeRole", err)

	status, _, refreshToken, err := s.iam.Verify(s.ctx, fmt.Sprint(ctx.Value("token")))
	if err != nil {
		return out, fmt.Errorf("%s", "Error verify from AuthChangeRole")
	}

	if status {
		// валидируем токен и получаем валидный, но просроченный токен
		out.Token, err = s.iam.Refresh(s.ctx, refreshToken, in.Profile, in.Expire)
		if err != nil {
			return out, fmt.Errorf("%s", "Error refresh from AuthChangeRole")
		}
	}

	return out, err
}
