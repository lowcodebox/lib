package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

const storageKey = "logger.field_storage"

type fieldMap map[string]Field

// fieldStorage структура хранения полей в map.
type fieldStorage struct {
	fields fieldMap
	sync.RWMutex
}

func (s *fieldStorage) SetField(f Field) {
	s.Lock()
	s.fields[f.key] = f
	s.Unlock()
}

func (s *fieldStorage) External() []zap.Field {
	s.RLock()
	externalFields := make([]zap.Field, 0, len(s.fields))

	for _, v := range s.fields {
		externalFields = append(externalFields, v.External())
	}
	s.RUnlock()

	return externalFields
}

func newStorage() *fieldStorage {
	mp := make(fieldMap, 5)
	return &fieldStorage{fields: mp}
}

func getStorageFromCtx(ctx context.Context) (*fieldStorage, bool) {
	if mc, ok := ctx.Value(storageKey).(*fieldStorage); ok && mc != nil {
		return mc, false
	}

	return newStorage(), true
}

func withStorageContext(ctx context.Context, storage *fieldStorage) context.Context {
	return context.WithValue(ctx, storageKey, storage)
}

func withStorageFields(ctx context.Context, logger *zap.Logger) *zap.Logger {
	storage, _ := getStorageFromCtx(ctx)
	return logger.With(storage.External()...)
}
