package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/packages/logger"
	"github.com/ReneKroon/ttlcache"
	"go.uber.org/zap"

	"github.com/pkg/errors"
)

const cacheKeyPrefix = "cache."

var ErrorCacheNotInit = errors.New("cache is not initialized")
var ErrorCachePersistentNotInit = errors.New("persistent cache is not initialized")
var ErrorKeyNotFound = errors.New("key is not found")
var ErrorValueNotCase = errors.New("failed to cast value")
var ErrorItemExpired = errors.New("item expired")

var (
	cacheCollection cache
	isCacheInit     bool // Добавляем флаг инициализации
)

type cache struct {
	ctx             context.Context
	items           sync.Map
	expiredInterval time.Duration // Интервал, через который GС удалит запись
	runGCInterval   time.Duration // Интервал запуска GC
	ttlcache        *ttlcache.Cache
}

type cacheItem struct {
	// Getter определяет механизм получения данных от любого источника к/р поддерживает интерфейс
	reader          func() (res interface{}, err error)
	cache           *ttlcache.Cache
	persistentCache *ttlcache.Cache
	locks           locks
	refreshInterval time.Duration // Интервал обновления кеша
	expiredTime     time.Time     // Время протухания кеша (когда GС удалит запись)
}

// Init инициализировали глобальную переменную defaultCache
// expiredInterval - время жизни записи в кеше, после удаляется с GC (0 - не удаляется никогда)
func Init(ctx context.Context, expiredInterval, runGCInterval time.Duration) {
	ttlcache := ttlcache.NewCache()
	ttlcache.SkipTtlExtensionOnHit(true)

	d := cache{
		ctx:             ctx,
		items:           sync.Map{},
		runGCInterval:   runGCInterval,
		expiredInterval: expiredInterval,
		ttlcache:        ttlcache,
	}
	cacheCollection = d
	isCacheInit = true // Устанавливаем флаг при инициализации

	go d.gc()
}

func Cache() *cache {
	if !isCacheInit { // Используем флаг для проверки инициализации
		return nil
	}

	return &cacheCollection
}

// Upsert регистрируем новый/обновляем кеш (указываем функцию, кр будет возвращать нужное значение)
func (c *cache) Upsert(key string, source func() (res interface{}, err error), refreshInterval time.Duration) (result interface{}, err error) {
	actual, loaded := c.items.Load(key)
	if loaded {
		item, ok := actual.(*cacheItem)
		if !ok {
			return nil, ErrorValueNotCase
		}

		item.reader = source
		item.refreshInterval = refreshInterval
		result, err = c.updateCacheValue(key, item)
		return result, err
	}

	expiredTime := time.Now().Add(c.expiredInterval)
	if c.expiredInterval == 0 {
		expiredTime = time.UnixMicro(0)
	}
	ci := cacheItem{
		cache:           c.ttlcache,
		persistentCache: c.ttlcache,
		locks: locks{
			keys: sync.Map{},
		},
		reader:          source,
		refreshInterval: refreshInterval,
		expiredTime:     expiredTime,
	}

	c.items.Store(key, &ci)
	result, err = c.updateCacheValue(key, &ci)

	return result, err
}

// Delete удаляем значение из кеша
func (c *cache) Delete(key string) {
	c.items.Delete(key)
}

// Get возвращает текущее значение параметра в сервисе keeper.
// Нужно учитывать, что значения на время кешируются и обновляются с заданной периодичностью.
func (c *cache) Get(key string) (value interface{}, err error) {
	var item *cacheItem

	actual, loaded := c.items.Load(key)
	if !loaded {
		return nil, ErrorKeyNotFound
	}

	item, ok := actual.(*cacheItem)
	if !ok {
		return nil, ErrorValueNotCase
	}

	if item.cache == nil {
		return nil, ErrorCacheNotInit
	}

	if item.persistentCache == nil {
		return nil, ErrorCachePersistentNotInit
	}

	if cachedValue, ok := item.cache.Get(key); ok {
		return cachedValue, nil
	}

	// Если стоит блокировка, значит кто-то уже обновляет кеш. В этом случае
	// пытаемся отдать предыдущее значение.
	if item.locks.Get(key) {
		return c.tryToGetOldValue(key)
	}

	// запускаем обновление фоном
	go c.updateCacheValue(key, item)

	// отдаем старое значение
	return c.tryToGetOldValue(key)
}

// updateCacheValue обновление значений в кеше
// вариант 1 - значение не найдено. Первый из запросов блокирует за собой обновление (на самом деле
// может возникнуть ситуация когда несколько запросов поставят блокировку и начнут
// обновлять кеш - пока считаем это некритичным).
func (c *cache) updateCacheValue(key string, item *cacheItem) (result interface{}, err error) {
	item.locks.Set(key, true)
	defer item.locks.Set(key, false)

	result, err = item.reader()
	if err != nil {
		return nil, errors.Wrap(err, "could not get value from getter")
	}

	item.cache.SetWithTTL(key, result, item.refreshInterval)
	item.persistentCache.Set(key, result)
	item.expiredTime = time.Now().Add(c.expiredInterval)

	return result, nil
}

// tryToGetOldValue пытается получить старое значение, если в момент запроса на актуальном стоит блокировка.
func (c *cache) tryToGetOldValue(key string) (interface{}, error) {
	var item *cacheItem

	actual, loaded := c.items.Load(key)
	if !loaded {
		return nil, ErrorKeyNotFound
	}

	item, ok := actual.(*cacheItem)
	if !ok {
		return nil, ErrorValueNotCase
	}

	fnGetPersistentCacheValue := func() (interface{}, error) {
		if cachedValue, ok := item.persistentCache.Get(key); ok {
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

// locks выполняет функции блокировки при одновременном обновлении значений в кеше.
type locks struct {
	// keys хранит информацию о локах по каждому отдельному ключу.
	// Если значение установлено в true, в данный момент обновление кеша захвачено одной из горутин.
	keys sync.Map
}

// Get возвращает информацию о том идет ли в данный момент обновление конкретного ключа.
func (l *locks) Get(key string) bool {
	value, ok := l.keys.Load(key)
	if !ok {
		return false
	}

	valueBool, ok := value.(bool)
	if !ok {
		return false
	}
	return valueBool
}

// Set устанавливает блокировку на обновление конкретного ключа другими горутинами.
func (l *locks) Set(key string, value bool) {
	l.keys.Store(key, value)
}

// Balancer запускаем балансировщик реплик для сервисов
// опрашивает список реплик из пингов и если запущено меньше, добавляет реплику.
func (c *cache) gc() {
	var err error
	ticker := time.NewTicker(c.runGCInterval)

	defer ticker.Stop()
	defer func() {
		lib.Recover(c.ctx)
		if err != nil {
			logger.Error(c.ctx, "error gc", zap.Error(err))
		}
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			err = c.cleaner()
			if err != nil {
				logger.Error(c.ctx, "error deleted item (expired time)", zap.Error(err))
			}
			ticker = time.NewTicker(c.runGCInterval)
		}
	}
}

func (c *cache) cleaner() (err error) {
	c.items.Range(func(key, value any) bool {
		item, ok := value.(*cacheItem)
		if !ok {
			return false
		}

		if item.expiredTime != time.UnixMicro(0) && !item.expiredTime.After(time.Now()) {
			c.items.Delete(key)
		}

		return true
	})
	return nil
}
