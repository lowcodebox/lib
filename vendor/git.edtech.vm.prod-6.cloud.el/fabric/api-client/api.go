// api клиент к сервису
// поддерживает CircuitBreaker
// (при срабатывании отдает ошибку и блокирует дальнейшие попытки запросов на заданный интервал)

package api

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"git.edtech.vm.prod-6.cloud.el/packages/cache"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const headerRequestId = "X-Request-Id"
const headerServiceKey = "X-Service-Key"
const tokenInterval = 1 * time.Minute
const constOperationDelete = "delete"
const constOperationAdd = "add"

type api struct {
	url        string
	observeLog bool
	//cb                  *gobreaker.CircuitBreaker
	cacheUpdateInterval time.Duration
	domain              string
	projectKey          string
}

type Api interface {
	Obj
}

type Obj interface {
	Data(ctx context.Context, tpls, option, role, page, size string) (result models.ResponseData, err error)
	ObjGet(ctx context.Context, uids string) (result models.ResponseData, err error)
	ObjGetWithCache(ctx context.Context, uids string) (result *models.ResponseData, err error)
	ObjCreate(ctx context.Context, bodymap map[string]string) (result models.ResponseData, err error)
	ObjDelete(ctx context.Context, uids string) (result models.ResponseData, err error)
	ObjAttrUpdate(ctx context.Context, uid, name, value, src, editor string) (result models.ResponseData, err error)
	LinkGet(ctx context.Context, tpl, obj, mode, short string) (result models.ResponseData, err error)
	LinkAdd(ctx context.Context, element, from, to string) (result models.ResponseData, err error)
	LinkDelete(ctx context.Context, element, from, to string) (result models.ResponseData, err error)
	LinkGetWithCache(ctx context.Context, tpl, obj, mode, short string) (result models.ResponseData, err error)
	Search(ctx context.Context, query, method, bodyJSON string) (resp string, err error)
	SearchWithCache(ctx context.Context, query, method, bodyJSON string) (resp string, err error)
	Tpls(ctx context.Context, role, option string) (result models.ResponseData, err error)
	Query(ctx context.Context, query, method, bodyJSON string) (result string, err error)
	QueryWithCache(ctx context.Context, query, method, bodyJSON string) (result string, err error)
	Element(ctx context.Context, action, body string) (result models.ResponseData, err error)
	ElementWithCache(ctx context.Context, action, body string) (result models.ResponseData, err error)
}

// Data - получение объектов по шаблону. Параметр option опциональный
func (a *api) Data(ctx context.Context, tpls, option, role, page, size string) (result models.ResponseData, err error) {
	//_, err = o.cb.Execute(func() (interface{}, error) {
	result, err = a.data(ctx, tpls, option, role, page, size)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error data primary haproxy", zap.Error(err))
		return result, fmt.Errorf("error request data (primary route) err: %w", err)
	}

	return result, err
}

func (a *api) Tpls(ctx context.Context, role, option string) (result models.ResponseData, err error) {
	//_, err = o.cb.Execute(func() (interface{}, error) {
	result, err = a.tpls(ctx, role, option)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error tpls primary haproxy", zap.Error(err))
		return result, fmt.Errorf("error request tpls (primary route) err: %w", err)
	}

	return result, err
}

// Search результат выводим в объект как при вызове Curl
func (a *api) Search(ctx context.Context, query, method, bodyJSON string) (resp string, err error) {
	resp, err = a.search(ctx, query, method, bodyJSON)
	if err != nil {
		logger.Error(ctx, "error UpdateFilter primary haproxy", zap.Error(err))
		return resp, fmt.Errorf("error request Search (primary route). err: %s", err)
	}

	return resp, nil
}

// SearchWithCache результат выводим в объект как при вызове Curl
// (с кешем если задан TTL кеширования при инициализации кеша)
func (a *api) SearchWithCache(ctx context.Context, query, method, bodyJSON string) (result string, err error) {

	//return a.Search(ctx, query, method, bodyJSON)

	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "QueryWithCache", err, query, method, bodyJSON)
	}
	key := lib.Hash(fmt.Sprintf("%s%s%s", query, method, bodyJSON))

	cacheValue, err := cache.Cache().Get(key)

	if errors.Is(err, cache.ErrorKeyNotFound) {
		var value interface{}
		value, err = cache.Cache().Upsert(key, func() (res interface{}, err error) {
			res, err = a.Search(ctx, query, method, bodyJSON)
			return res, err
		}, a.cacheUpdateInterval)
		if err == nil && value != nil {
			return fmt.Sprint(value), nil
		} else {
			err = fmt.Errorf("error exec cache query (Query). err: %s, value is empty: %t, value: %+v", err, value == nil, value)
		}
	}

	if err != nil {
		logger.Error(ctx, "error exec cache query", zap.String("func", "QueryWithCache"), zap.String("key", key), zap.Error(err))
		cacheValue, err = a.Search(ctx, query, method, bodyJSON)
		if err != nil {
			return result, fmt.Errorf("error get cache (Query). err: %s", err)
		}
	}

	return fmt.Sprint(cacheValue), err
}

// Query результат выводим в объект как при вызове Curl
func (a *api) Query(ctx context.Context, query, method, bodyJSON string) (result string, err error) {
	//_, err = a.cb.Execute(func() (interface{}, error) {
	result, err = a.query(ctx, query, method, bodyJSON)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error UpdateFilter primary haproxy", zap.Error(err))
		return "", fmt.Errorf("error request Query (primary route). err: %s", err)
	}

	return result, err
}

// QueryWithCache результат выводим в объект как при вызове Curl
// (с кешем если задан TTL кеширования при инициализации кеша)
func (a *api) QueryWithCache(ctx context.Context, query, method, bodyJSON string) (result string, err error) {

	//return a.Query(ctx, query, method, bodyJSON)

	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "QueryWithCache", err, query, method, bodyJSON)
	}
	key := lib.Hash(fmt.Sprintf("%s%s%s", query, method, bodyJSON))

	cacheValue, err := cache.Cache().Get(key)

	if errors.Is(err, cache.ErrorKeyNotFound) {
		var value interface{}
		value, err = cache.Cache().Upsert(key, func() (res interface{}, err error) {
			res, err = a.Query(ctx, query, method, bodyJSON)
			return res, err
		}, a.cacheUpdateInterval)
		if err == nil && value != nil {
			return fmt.Sprint(value), nil
		} else {
			err = fmt.Errorf("error exec cache query (Query). err: %s, value is empty: %t, value: %+v", err, value == nil, value)
		}
	}

	if err != nil {
		logger.Error(ctx, "error exec cache query", zap.String("func", "QueryWithCache"), zap.String("key", key), zap.Error(err))
		cacheValue, err = a.Query(ctx, query, method, bodyJSON)
		if err != nil {
			return result, fmt.Errorf("error get cache (Query). err: %s", err)
		}
	}

	return fmt.Sprint(cacheValue), err
}

func (a *api) ObjGet(ctx context.Context, uids string) (result models.ResponseData, err error) {
	if uids == "" {
		return result, fmt.Errorf("error ObjGet. uids is empty")
	}
	//_, err = a.cb.Execute(func() (interface{}, error) {
	result, err = a.objGet(ctx, uids)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error ObjGet primary haproxy", zap.Error(err))
		return result, fmt.Errorf("error request ObjGet (primary route). err: %s", err)
	}

	return result, err
}

func (a *api) ObjGetWithCache(ctx context.Context, uids string) (result *models.ResponseData, err error) {
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "ObjGetWithCache", err, uids)
	}
	key := lib.Hash(uids)

	cacheValue, err := cache.Cache().Get(key)

	if errors.Is(err, cache.ErrorKeyNotFound) {
		var value interface{}
		value, err = cache.Cache().Upsert(key, func() (res interface{}, err error) {
			res, err = a.ObjGet(ctx, uids)
			return res, err
		}, a.cacheUpdateInterval)
		if err == nil && value != nil {
			res, ok := value.(models.ResponseData)
			if ok {
				return &res, nil
			}
			err = fmt.Errorf("error cast type in cache query (ObjGetWithCache). err: %s, value is empty: %t, result: %+v", err, value == nil, value)
		} else {
			err = fmt.Errorf("error exec cache query (ObjGet). err: %s, value is empty: %t", err, value == nil)
		}
	}

	// повторяем запрос (без кеша)
	if err != nil {
		logger.Error(ctx, "error exec cache query", zap.String("func", "ObjGetWithCache"), zap.String("key", key), zap.Error(err))
		cacheValue, err = a.ObjGet(ctx, uids)
		if err == nil && cacheValue != nil {
			res, ok := cacheValue.(models.ResponseData)
			if ok {
				return &res, nil
			}
			err = fmt.Errorf("error cast type in query (ObjGet). err: %s, value is empty: %t, result: %+v", err, cacheValue == nil, cacheValue)
		} else {
			err = fmt.Errorf("error exec query (ObjGet). err: %s, value is empty: %t", err, cacheValue == nil)
		}
	}

	if err == nil && cacheValue != nil {
		res, ok := cacheValue.(models.ResponseData)
		if ok {
			return &res, nil
		}
		err = fmt.Errorf("error cast type in query (ObjGet). err: %s, value is empty: %t, result: %+v", err, cacheValue == nil, cacheValue)
	} else {
		err = fmt.Errorf("error exec query (ObjGet). err: %s, value is empty: %t", err, cacheValue == nil)
	}

	return result, err
}

// LinkDelete - удаление линки
func (a *api) LinkDelete(ctx context.Context, element, from, to string) (result models.ResponseData, err error) {
	//_, err = a.cb.Execute(func() (interface{}, error) {
	result, err = a.linkOperation(ctx, constOperationDelete, element, from, to)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error LinkDelete primary haproxy",
			zap.String("operation", constOperationDelete),
			zap.String("element", element),
			zap.String("from", from),
			zap.String("to", to))
		//zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request LinkDelete (primary route). check apiCircuitBreaker. err: %s", err)
	}

	return result, err
}

// LinkAdd - добавление линки в объект
func (a *api) LinkAdd(ctx context.Context, element, from, to string) (result models.ResponseData, err error) {
	//_, err = a.cb.Execute(func() (interface{}, error) {
	result, err = a.linkOperation(ctx, constOperationAdd, element, from, to)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error LinkAdd primary haproxy",
			zap.String("operation", constOperationAdd),
			zap.String("element", element),
			zap.String("from", from),
			zap.String("to", to))
		//zap.Any("status CircuitBreaker", a.cb.State().String()), zap.Error(err))
		return result, fmt.Errorf("error request LinkAdd (primary route). check apiCircuitBreaker. err: %s", err)
	}

	return result, err
}

// LinkGet - получение связанных объектов
func (a *api) LinkGet(ctx context.Context, tpl, obj, mode, short string) (result models.ResponseData, err error) {
	//_, err = a.cb.Execute(func() (interface{}, error) {
	result, err = a.linkGet(ctx, tpl, obj, mode, short)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error LinkGet primary haproxy", zap.Error(err))
		return result, fmt.Errorf("error request LinkGet (primary route). err: %s", err)
	}

	return result, err
}

// LinkGetWithCache - получение связанных объектов
// (с кешем если задан TTL кеширования при инициализации кеша)
func (a *api) LinkGetWithCache(ctx context.Context, tpl, obj, mode, short string) (result models.ResponseData, err error) {
	var ok bool
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "LinkGetWithCache", err, tpl, obj, mode, short)
	}
	key := lib.Hash(fmt.Sprintf("%s%s%s%s", tpl, obj, mode, short))

	cacheValue, err := cache.Cache().Get(key)

	if errors.Is(err, cache.ErrorKeyNotFound) {
		var value interface{}
		value, err = cache.Cache().Upsert(key, func() (res interface{}, err error) {
			res, err = a.LinkGet(ctx, tpl, obj, mode, short)
			return res, err
		}, a.cacheUpdateInterval)
		if err == nil && value != nil {
			result, ok = value.(models.ResponseData)
			if ok {
				return result, nil
			}
			err = fmt.Errorf("error cast type in cache query (LinkGetWithCache). err: %s, value is empty: %t, result: %+v", err, value == nil, value)
		} else {
			err = fmt.Errorf("error exec cache query (LinkGet). err: %s, value is empty: %t", err, value == nil)
		}
	}

	// повторяем запрос (без кеша)
	if err != nil {
		logger.Error(ctx, "error exec cache query", zap.String("func", "LinkGetWithCache"), zap.String("key", key), zap.Error(err))
		cacheValue, err = a.LinkGet(ctx, tpl, obj, mode, short)
		if err == nil && cacheValue != nil {
			res, ok := cacheValue.(models.ResponseData)
			if ok {
				return res, nil
			}
			err = fmt.Errorf("error cast type in query (LinkGet). err: %s, value is empty: %t, result: %+v", err, cacheValue == nil, cacheValue)
		} else {
			err = fmt.Errorf("error exec query (LinkGet). err: %s, value is empty: %t", err, cacheValue == nil)
		}
	}

	if err == nil && cacheValue != nil {
		res, ok := cacheValue.(models.ResponseData)
		if ok {
			return res, nil
		}
		err = fmt.Errorf("error cast type in query (ObjGet). err: %s, value is empty: %t, result: %+v", err, cacheValue == nil, cacheValue)
	} else {
		err = fmt.Errorf("error exec query (ObjGet). err: %s, value is empty: %t", err, cacheValue == nil)
	}

	return result, err
}

// ObjAttrUpdate изменение значения аттрибута объекта
func (a *api) ObjAttrUpdate(ctx context.Context, uid, name, value, src, editor string) (result models.ResponseData, err error) {
	if uid == "" {
		return result, fmt.Errorf("[ObjAttrUpdate] error ObjAttrUpdate. uid is empty")
	}
	//_, err = a.cb.Execute(func() (interface{}, error) {
	result, err = a.objAttrUpdate(ctx, uid, name, value, src, editor)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error ObjAttrUpdate primary haproxy", zap.Error(err))
		return result, fmt.Errorf("[ObjAttrUpdate] error request ObjAttrUpdate (primary route). err: %s", err)
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
	//_, err = a.cb.Execute(func() (interface{}, error) {
	result, err = a.element(ctx, action, body)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error Element primary haproxy", zap.Error(err))
		return result, fmt.Errorf("[Element] error request Element (primary route). err: %s", err)
	}

	return result, err
}

// ElementWithCache - операции с полями для объекта
// кешируем только операции получения данных, остальные без кеша
// (с кешем если задан TTL кеширования при инициализации кеша)
func (a *api) ElementWithCache(ctx context.Context, action, body string) (result models.ResponseData, err error) {

	//return a.Element(ctx, action, body)

	var ok bool
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "ElementWithCache", err, action, body)
	}
	key := lib.Hash(fmt.Sprintf("%s%s", action, body))

	if action != "elements" && action != "all" {
		return a.Element(ctx, action, body)
	}

	cacheValue, err := cache.Cache().Get(key)

	if errors.Is(err, cache.ErrorKeyNotFound) {
		var value interface{}
		value, err = cache.Cache().Upsert(key, func() (res interface{}, err error) {
			res, err = a.Element(ctx, action, body)
			return res, err
		}, a.cacheUpdateInterval)
		if err == nil && value != nil {
			result, ok = value.(models.ResponseData)
			if ok {
				return result, nil
			}
			err = fmt.Errorf("error cast type in cache query (ElementWithCache). err: %s, value is empty: %t, result: %+v", err, value == nil, value)
		} else {
			err = fmt.Errorf("error exec cache query. err: %s, value is empty: %t", err, value == nil)
		}
	}

	// повторяем запрос (без кеша)
	if err != nil {
		logger.Error(ctx, "error exec cache query", zap.String("func", "ElementWithCache"), zap.String("key", key), zap.Error(err))
		cacheValue, err = a.Element(ctx, action, body)
		if err == nil && cacheValue != nil {
			res, ok := cacheValue.(models.ResponseData)
			if ok {
				return res, nil
			}
			err = fmt.Errorf("error cast type in query (Element). err: %s, value is empty: %t, result: %+v", err, cacheValue == nil, cacheValue)
		} else {
			err = fmt.Errorf("error exec query (Element). err: %s, value is empty: %t", err, cacheValue == nil)
		}
	}

	if err == nil && cacheValue != nil {
		res, ok := cacheValue.(models.ResponseData)
		if ok {
			return res, nil
		}
		err = fmt.Errorf("error cast type in query (ObjGet). err: %s, value is empty: %t, result: %+v", err, cacheValue == nil, cacheValue)
	} else {
		err = fmt.Errorf("error exec query (ObjGet). err: %s, value is empty: %t", err, cacheValue == nil)
	}

	return result, err
}

func (a *api) ObjCreate(ctx context.Context, bodymap map[string]string) (result models.ResponseData, err error) {
	//_, err = a.cb.Execute(func() (interface{}, error) {
	result, err = a.objCreate(ctx, bodymap)
	//if err != nil {
	//	err = fmt.Errorf("error ObjCreate, bodymap: %+v, err: %s)", bodymap, err)
	//}
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error ObjCreate primary haproxy", zap.Error(err), zap.Any("bodymap", bodymap))
		return result, fmt.Errorf("error request ObjCreate (primary route). err: %s", err)
	}

	return result, err
}

func (a *api) ObjDelete(ctx context.Context, uids string) (result models.ResponseData, err error) {
	//_, err = a.cb.Execute(func() (interface{}, error) {
	result, err = a.objDelete(ctx, uids)
	//return result, err
	//})
	if err != nil {
		logger.Error(ctx, "error ObjDelete primary haproxy", zap.Error(err))
		return result, fmt.Errorf("error request ObjDelete (primary route). err: %s", err)
	}

	return result, err
}

func New(ctx context.Context, urlstr string, observeLog bool, cacheUpdateInterval time.Duration, cbMaxRequests uint32, cbTimeout, cbInterval time.Duration, projectKey string) Api {
	//var err error
	if cbMaxRequests == 0 {
		cbMaxRequests = 3
	}
	if cbTimeout == 0 {
		cbTimeout = 5 * time.Second
	}
	if cbInterval == 0 {
		cbInterval = 5 * time.Second
	}

	//cb := gobreaker.NewCircuitBreaker(
	//	gobreaker.Settings{
	//		Name:        "apiCircuitBreaker",
	//		MaxRequests: cbMaxRequests, // максимальное количество запросов, которые могут пройти, когда автоматический выключатель находится в полуразомкнутом состоянии
	//		Timeout:     cbTimeout,     // период разомкнутого состояния, после которого выключатель переходит в полуразомкнутое состояние
	//		Interval:    cbInterval,    // циклический период замкнутого состояния автоматического выключателя для сброса внутренних счетчиков
	//		ReadyToTrip: func(counts gobreaker.Counts) bool {
	//			logger.Error(ctx, "apiCircuitBreaker is ReadyToTrip", zap.Any("counts.ConsecutiveFailures", counts.ConsecutiveFailures), zap.Error(err))
	//			return counts.ConsecutiveFailures > 2
	//		},
	//		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
	//			logger.Error(ctx, "apiCircuitBreaker changed position", zap.Any("name", name), zap.Any("from", from), zap.Any("to", to), zap.Error(err))
	//		},
	//	},
	//)

	u, _ := url.Parse(urlstr)
	splitUrl := strings.Split(u.Path, "/")

	domain := splitUrl[:]

	// инициализировали переменную кеша
	cache.Init(ctx, 10*time.Hour, 10*time.Minute)

	return &api{
		urlstr,
		observeLog,
		cacheUpdateInterval,
		strings.Join(domain, "/"),
		projectKey,
	}
}
