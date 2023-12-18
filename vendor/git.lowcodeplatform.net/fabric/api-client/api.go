// api клиент к сервису
// поддерживает CircuitBreaker
// (при срабатывании отдает ошибку и блокирует дальнейшие попытки запросов на заданный интервал)

package api

import (
	"context"
	"fmt"
	"time"

	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"

	"github.com/sony/gobreaker"
)

const headerRequestId = "X-Request-Id"

type api struct {
	url        string
	observeLog bool
	cb         *gobreaker.CircuitBreaker
}

type Api interface {
	Obj
}

type Obj interface {
	ObjGet(ctx context.Context, uids string) (result models.ResponseData, err error)
	ObjCreate(ctx context.Context, bodymap map[string]string) (result models.ResponseData, err error)
	ObjDelete(ctx context.Context, uids string) (result models.ResponseData, err error)
	ObjAttrUpdate(ctx context.Context, uid, name, value, src, editor string) (result models.ResponseData, err error)
	LinkGet(ctx context.Context, tpl, obj, mode, short string) (result models.ResponseData, err error)
	Query(ctx context.Context, query, method, bodyJSON string) (result string, err error)
	Element(ctx context.Context, action, body string) (result models.ResponseData, err error)
}

// Query результат выводим в объект как при вызове Curl
func (a *api) Query(ctx context.Context, query, method, bodyJSON string) (result string, err error) {
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.query(ctx, query, method, bodyJSON)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error UpdateFilter primary haproxy", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return "", fmt.Errorf("error request Query (primary route). check apiCircuitBreaker. err: %s", err)
	}

	return result, err
}

func (a *api) ObjGet(ctx context.Context, uids string) (result models.ResponseData, err error) {
	if uids == "" {
		return result, fmt.Errorf("error ObjGet. uids is empty")
	}
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.objGet(ctx, uids)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error ObjGet primary haproxy", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request ObjGet (primary route). check apiCircuitBreaker. err: %s", err)
	}

	return result, err
}

func (a *api) LinkGet(ctx context.Context, tpl, obj, mode, short string) (result models.ResponseData, err error) {
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.linkGet(ctx, tpl, obj, mode, short)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error LinkGet primary haproxy", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request LinkGet (primary route). check apiCircuitBreaker. err: %s", err)
	}

	return result, err
}

// ObjAttrUpdate изменение значения аттрибута объекта
func (a *api) ObjAttrUpdate(ctx context.Context, uid, name, value, src, editor string) (result models.ResponseData, err error) {
	if uid == "" {
		return result, fmt.Errorf("error ObjAttrUpdate. uid is empty")
	}
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.objAttrUpdate(ctx, uid, name, value, src, editor)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error ObjAttrUpdate primary haproxy", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request ObjAttrUpdate (primary route). check apiCircuitBreaker. err: %s", err)
	}

	return result, err
}

// Element
// TODO ПЕРЕДЕЛАТЬ на понятные пути в ORM
// сделано так для совместимости со старой версией GUI
// Action:
// block - Блокировка поля для его редактировании
// unblock - Разблокировка заблокированного поля
// update - Обновление значения заблокированного поля
// checkup - Проверяем переданное значение на соответствие выбранному условию
// all (elements) - Получаем поля, по заданному в параметрах типу
// "" - без действия - получаем все поля для объекта
func (a *api) Element(ctx context.Context, action, body string) (result models.ResponseData, err error) {
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.element(ctx, action, body)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error Element primary haproxy", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request Element (primary route). check apiCircuitBreaker. err: %s", err)
	}

	return result, err
}

func (a *api) ObjCreate(ctx context.Context, bodymap map[string]string) (result models.ResponseData, err error) {
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.objCreate(ctx, bodymap)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error ObjCreate primary haproxy", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request ObjCreate (primary route). check apiCircuitBreaker. err: %s", err)
	}

	return result, err
}

func (a *api) ObjDelete(ctx context.Context, uids string) (result models.ResponseData, err error) {
	_, err = a.cb.Execute(func() (interface{}, error) {
		result, err = a.objDelete(ctx, uids)
		return result, err
	})
	if err != nil {
		logger.Error(ctx, "error ObjDelete primary haproxy", zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request ObjDelete (primary route). check apiCircuitBreaker. err: %s", err)
	}

	return result, err
}

func New(ctx context.Context, url string, observeLog bool, cbMaxRequests uint32, cbTimeout, cbInterval time.Duration) Api {
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
			Name:        "apiCircuitBreaker",
			MaxRequests: cbMaxRequests, // максимальное количество запросов, которые могут пройти, когда автоматический выключатель находится в полуразомкнутом состоянии
			Timeout:     cbTimeout,     // период разомкнутого состояния, после которого выключатель переходит в полуразомкнутое состояние
			Interval:    cbInterval,    // циклический период замкнутого состояния автоматического выключателя для сброса внутренних счетчиков
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				logger.Error(ctx, "apiCircuitBreaker is ReadyToTrip", zap.Any("counts.ConsecutiveFailures", counts.ConsecutiveFailures), zap.Error(err))
				return counts.ConsecutiveFailures > 2
			},
			OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
				logger.Error(ctx, "apiCircuitBreaker changed position", zap.Any("name", name), zap.Any("from", from), zap.Any("to", to), zap.Error(err))
			},
		},
	)

	return &api{
		url,
		observeLog,
		cb,
	}
}