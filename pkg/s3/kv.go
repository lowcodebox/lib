package s3

import (
	"context"
	"fmt"
	"sync"
)

type KVStore interface {
	Check(ctx context.Context, key string) (ok bool, err error)
	Put(ctx context.Context, key string, val string) (err error)
	Get(ctx context.Context, key string) (val string, err error)
}

type LocalKVStore struct {
	vals    map[string]string
	rwMutex sync.RWMutex
}

func (b *LocalKVStore) Check(ctx context.Context, key string) (ok bool, err error) {
	if _, ok := b.vals[key]; ok {
		return true, nil
	}
	return false, nil
}

func (b *LocalKVStore) Put(_ context.Context, key string, val string) (err error) {
	b.rwMutex.Lock()
	defer b.rwMutex.Unlock()
	b.vals[key] = val
	return nil
}

func (b *LocalKVStore) Get(ctx context.Context, key string) (val string, err error) {
	b.rwMutex.RLock()
	defer b.rwMutex.RUnlock()
	ok, err := b.Check(ctx, key)
	if err != nil {
		return "", err
	}
	if ok {
		return b.vals[key], nil
	}
	return "", fmt.Errorf("missing required field: %s", key)
}

func NewLocalKVStore() KVStore {
	return &LocalKVStore{
		vals:    make(map[string]string),
		rwMutex: sync.RWMutex{},
	}
}

func InitializeWithMap(ctx context.Context, store KVStore, initial map[ConfigField]string) (err error) {
	for k, v := range initial {
		err = store.Put(ctx, string(k), v)
		if err != nil {
			return err
		}
	}
	return nil
}
