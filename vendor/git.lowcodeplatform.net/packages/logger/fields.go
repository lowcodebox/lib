package logger

import (
	"context"
	"sync"
)

const (
	requestIDField 	string = "request-id"
	serviceIDField 	string = "service-id"
	configIDField 	string = "config-id"
)

//nolint:gochecknoglobals
var (
	logKeys = make(map[key]struct{})
	mtx     sync.RWMutex
)

type key string

func SetFieldCtx(ctx context.Context, name, val string) context.Context {
	nameKey := key("logger." + name)

	mtx.RLock()
	_, ok := logKeys[nameKey]
	mtx.RUnlock()

	if !ok {
		mtx.Lock()
		logKeys[nameKey] = struct{}{}
		mtx.Unlock()
	}
	
	return context.WithValue(ctx, nameKey, val)
}

// WithFieldsContext получение контекста из полей явного типа,
// при логировании поля будут прописаны в external логгер (zap).
func WithFieldsContext(ctx context.Context, fields ...Field) context.Context {
	storage, isNew := getStorageFromCtx(ctx)
	for _, field := range fields {
		storage.SetField(field)
	}

	if isNew {
		// если в контексте не был записан storage возвращаем с родительским контекстом
		return withStorageContext(ctx, storage)
	}
	// если он уже есть, достаточно вернуть текущий контекст, тк storage ссылочный
	return ctx
}

func GetFieldCtx(ctx context.Context, name string) string {
	nameKey := key("logger." + name)
	requestID, _ := ctx.Value(nameKey).(string)

	return requestID
}

// SetRequestIDCtx sets request-id log field via context.
// Tip: use HTTPMiddleware instead.
func SetRequestIDCtx(ctx context.Context, val string) context.Context {
	return SetFieldCtx(ctx, requestIDField, val)
}

func GetRequestIDCtx(ctx context.Context) string {
	return GetFieldCtx(ctx, requestIDField)
}

func SetServiceIDCtx(ctx context.Context, val string) context.Context {
	return SetFieldCtx(ctx, serviceIDField, val)
}

func GetServiceIDCtx(ctx context.Context) string {
	return GetFieldCtx(ctx, serviceIDField)
}

func SetConfigIDCtx(ctx context.Context, val string) context.Context {
	return SetFieldCtx(ctx, configIDField, val)
}

func GetConfigIDCtx(ctx context.Context) string {
	return GetFieldCtx(ctx, configIDField)
}