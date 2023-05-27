package lib

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ReneKroon/ttlcache"

	"github.com/pkg/errors"
)

const cacheKeyPrefix = "cache."

var (
	defaultCache *cache
)

// locks выполняет функции блокировки при одновременном обновлении значений в кеше.
type locks struct {
	// keys хранит информацию о локах по каждому отдельному ключу.
	// Если значение установлено в true, в данный момент обновление кеша захвачено одной из горутин.
	keys map[string]bool
	mx   sync.RWMutex
}

// Get возвращает информацию о том идет ли в данный момент обновление конкретного ключа.
func (l *locks) Get(key string) bool {
	l.mx.RLock()
	defer l.mx.RUnlock()

	return l.keys[key]
}

// Set устанавливает блокировку на обновление конкретного ключа другими горутинами.
func (l *locks) Set(key string, value bool) {
	l.mx.Lock()
	l.keys[key] = value
	l.mx.Unlock()
}

type cache struct {
	// Getter определяет механизм получения данных от любого источника к/р поддерживает интерфейс
	reader          io.Reader
	cache           *ttlcache.Cache
	persistentCache *ttlcache.Cache
	locks           locks
	cacheTTL        time.Duration
}

// Get возвращает текущее значение параметра в сервисе keeper.
// Нужно учитывать, что значения на время кешируются и обновляются с заданной периодичностью.
func (c *cache) Get(key string) (value interface{}, err error) {
	if c.cache == nil {
		return nil, fmt.Errorf("cache is not inited")
	}

	if c.persistentCache == nil {
		return nil, fmt.Errorf("persistent cache is not inited")
	}

	if cachedValue, ok := c.cache.Get(cacheKeyPrefix + key); ok {
		return cachedValue, nil
	}

	// Если стоит блокировка, значит кто-то уже обновляет кеш. В этом случае
	// пытаемся отдать предыдущее значение.
	if c.locks.Get(key) {
		return c.tryToGetOldValue(key)
	}

	// Значение не найдено. Первый из запросов блокирует за собой обновление (на самом деле
	// может возникнуть ситуация когда несколько запросов поставят блокировку и начнут
	// обновлять кеш - пока считаем это некритичным).
	c.locks.Set(key, true)
	defer c.locks.Set(key, false)

	values := []byte{}
	_, err = c.reader.Read(values)
	if err != nil {
		return nil, errors.Wrap(err, "could not get value from getter")
	}

	c.cache.SetWithTTL(cacheKeyPrefix+key, value, c.cacheTTL)
	c.persistentCache.Set(cacheKeyPrefix+key, value)

	return value, nil
}

// tryToGetOldValue пытается получить старое значение, если в момент запроса на актуальном стоит блокировка.
func (c *cache) tryToGetOldValue(key string) (interface{}, error) {
	fnGetPersistentCacheValue := func() (interface{}, error) {
		if cachedValue, ok := c.persistentCache.Get(cacheKeyPrefix + key); ok {
			return cachedValue, nil
		}

		return nil, fmt.Errorf("persinstent cache is empty")
	}

	oldValue, err := fnGetPersistentCacheValue()

	// Повторяем попытку получить значение. При старте сервиса может возникнуть блокировка
	// обновления ключа, но при этом в постоянном кеше еще может не быть значения.
	if err != nil {
		time.Sleep(100 * time.Millisecond)

		oldValue, err = fnGetPersistentCacheValue()
	}

	return oldValue, err
}

// CacheInit инициализировали глобальную переменную defaultCache
// source - источник, откуда мы получаем значения для кеширования
func CacheInit(ttl time.Duration, source io.Reader) {
	defaultCache = &cache{
		cacheTTL: ttl,
		reader:   source,
	}
}
