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

var ErrorCacheNotInit = errors.New("cache is not inited")
var ErrorCachePersistentNotInit = errors.New("persistent cache is not inited")
var ErrorKeyNotFound = errors.New("key is not found")
var ErrorItemExpired = errors.New("item expired")

var (
	cacheCollection cache
)

type cache struct {
	ctx             context.Context
	items           map[string]*cacheItem
	mx              sync.RWMutex
	expiredInterval time.Duration // интервал, через который GС удалит запись
	runGCInterval   time.Duration // интервал запуска GC
}

type cacheItem struct {
	// Getter определяет механизм получения данных от любого источника к/р поддерживает интерфейс
	reader          func() (res interface{}, err error)
	cache           *ttlcache.Cache
	persistentCache *ttlcache.Cache
	locks           locks
	refreshInterval time.Duration // интервал обновления кеша
	expiredTime     time.Time     // время протухания кеша (когда GС удалит запись)
}

func Cache() *cache {
	if &cacheCollection == nil {
		panic("cache has not been initialized, call CacheRegister() before use")
	}

	return &cacheCollection
}

// Upsert регистрируем новый/обновляем кеш (указываем фукнцию, кр будет возвращать нужное значение)
func (c *cache) Upsert(key string, source func() (res interface{}, err error), refreshInterval time.Duration) (err error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.items == nil {
		c.items = map[string]*cacheItem{}
	}

	item, found := c.items[key]
	if found {
		item.reader = source
		item.refreshInterval = refreshInterval
		_, err = c.updateCacheValue(key, item)
		return err
	}

	cache := ttlcache.NewCache()
	cache.SkipTtlExtensionOnHit(true)

	expiredTime := time.Now().Add(c.expiredInterval)
	if c.expiredInterval == 0 {
		expiredTime = time.UnixMicro(0)
	}
	ci := cacheItem{
		cache:           cache,
		persistentCache: ttlcache.NewCache(),
		locks:           locks{keys: map[string]bool{}},
		reader:          source,
		refreshInterval: refreshInterval,
		expiredTime:     expiredTime,
	}

	c.items[key] = &ci

	return err
}

// Delete удаляем значение из кеша
func (c *cache) Delete(key string) (err error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	delete(c.items, key)
	return err
}

// Get возвращает текущее значение параметра в сервисе keeper.
// Нужно учитывать, что значения на время кешируются и обновляются с заданной периодичностью.
func (c *cache) Get(key string) (value interface{}, err error) {
	var item *cacheItem
	var found bool

	item, found = c.items[key]
	if !found {
		return nil, ErrorKeyNotFound
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
	var found bool

	item, found = c.items[key]
	if !found {
		return nil, fmt.Errorf("error. key is not found")
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

// Init инициализировали глобальную переменную defaultCache
// expiredInterval - время жизни записи в кеше, после удаляется с GC (0 - не удаляется никогда)
func Init(ctx context.Context, expiredInterval, runGCInterval time.Duration) {
	d := cache{
		ctx:             ctx,
		items:           map[string]*cacheItem{},
		mx:              sync.RWMutex{},
		runGCInterval:   runGCInterval,
		expiredInterval: expiredInterval,
	}
	cacheCollection = d

	go d.gc()
}

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
	c.mx.Lock()
	defer c.mx.Unlock()

	// удаляем значение ключа из хранилища (чтобы не копились старые хеши)
	for key, item := range c.items {
		if item.expiredTime != time.UnixMicro(0) && !item.expiredTime.After(time.Now()) {
			err = c.Delete(key)
			if err != nil {
				return err
			}
		}
	}

	return err
}
