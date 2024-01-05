package client

import (
	"context"
	"fmt"
	"time"

	"git.lowcodeplatform.net/fabric/iam/pkg/i18n"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"

	"github.com/sony/gobreaker"
)

type iam struct {
	ctx        context.Context
	url        string
	projectKey string
	msg        i18n.I18n
	observeLog bool
	cb         *gobreaker.CircuitBreaker
}

type IAM interface {
	Verify(ctx context.Context, tokenString string) (statue bool, body *models.Token, refreshToken string, err error)
	Refresh(ctx context.Context, token, profile string, expire bool) (result string, err error)
	ProfileGet(ctx context.Context, sessionID string) (result string, err error)
	ProfileList(ctx context.Context) (result string, err error)
}

func (a *iam) Refresh(ctx context.Context, token, profile string, expire bool) (result string, err error) {
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.refresh(ctx, token, profile, expire)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error Refresh primary iam", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request Refresh (primary route). check iamCircuitBreaker. err: %s", err)
	}

	return result, err
}

func (a *iam) ProfileGet(ctx context.Context, sessionID string) (result string, err error) {
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.profileGet(ctx, sessionID)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error ProfileGet primary iam", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request ProfileGet (primary route). check iamCircuitBreaker. err: %s", err)
	}

	return result, err
}

func (a *iam) ProfileList(ctx context.Context) (result string, err error) {
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.profileList(ctx)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error ProfileList primary iam", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request ProfileList (primary route). check iamCircuitBreaker. err: %s", err)
	}

	return result, err
}

func (a *iam) Verify(ctx context.Context, tokenString string) (status bool, body *models.Token, refreshToken string, err error) {
	_, err = a.cb.Execute(func() (interface{}, error) {
		status, body, refreshToken, err = a.verify(ctx, tokenString)
		return status, err
	})
	if err != nil {
		logger.Error(ctx, "error Verify primary iam", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return status, body, refreshToken, fmt.Errorf("error request Verify (primary route). check iamCircuitBreaker. err: %s", err)
	}

	return status, body, refreshToken, err
}

func New(ctx context.Context, url, projectKey string, observeLog bool, cbMaxRequests uint32, cbTimeout, cbInterval time.Duration) IAM {
	if url[len(url)-1:] == "/" {
		url = url[:len(url)-1]
	}

	var err error
	if cbMaxRequests == 0 {
		cbMaxRequests = 3
	}
	if cbTimeout == 0 {
		cbTimeout = 5 * time.Second
	}
	if cbInterval == 0 {
		cbInterval = 5 * time.Second
	}

	cb := gobreaker.NewCircuitBreaker(
		gobreaker.Settings{
			Name:        "iamCircuitBreaker",
			MaxRequests: cbMaxRequests, // максимальное количество запросов, которые могут пройти, когда автоматический выключатель находится в полуразомкнутом состоянии
			Timeout:     cbTimeout,     // период разомкнутого состояния, после которого выключатель переходит в полуразомкнутое состояние
			Interval:    cbInterval,    // циклический период замкнутого состояния автоматического выключателя для сброса внутренних счетчиков
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				logger.Error(ctx, "iamCircuitBreaker is ReadyToTrip", zap.Any("counts.ConsecutiveFailures", counts.ConsecutiveFailures), zap.Error(err))
				return counts.ConsecutiveFailures > 2
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				logger.Error(ctx, "iamCircuitBreaker changed position", zap.Any("name", name), zap.Any("from", from), zap.Any("to", to), zap.Error(err))
			},
		},
	)

	msg := i18n.New()
	return &iam{
		ctx,
		url,
		projectKey,
		msg,
		observeLog,
		cb,
	}
}
