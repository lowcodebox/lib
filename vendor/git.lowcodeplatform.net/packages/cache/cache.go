package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/ReneKroon/ttlcache"

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
	items map[string]*cacheItem
	mx    sync.RWMutex
	ttl   time.Duration // интервал удаление записи из кеша
}

type cacheItem struct {
	// Getter определяет механизм получения данных от любого источника к/р поддерживает интерфейс
	reader          func() (res interface{}, err error)
	cache           *ttlcache.Cache
	persistentCache *ttlcache.Cache
	locks           locks
	cacheTTL        time.Duration // интервал протухания/обновления кеша
	expired         time.Time     // время удаляения записи из кеша
}

func Cache() *cache {
	if &cacheCollection == nil {
		panic("cache has not been initialized, call CacheRegister() before use")
	}

	return &cacheCollection
}

// Upsert регистрируем новый/обновляем кеш (указываем фукнцию, кр будет возвращать нужное значение)
func (c *cache) Upsert(key string, source func() (res interface{}, err error), ttl time.Duration) (err error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.items == nil {
		c.items = map[string]*cacheItem{}
	}

	item, found := c.items[key]
	if found {
		item.reader = source
		item.cacheTTL = ttl
		_, err = c.updateCacheValue(key, item)
		return err
	}

	cache := ttlcache.NewCache()
	cache.SkipTtlExtensionOnHit(true)

	expiredTime := time.Now().Add(c.ttl)
	if c.ttl == 0 {
		expiredTime = time.UnixMicro(0)
	}
	ci := cacheItem{
		cache:           cache,
		persistentCache: ttlcache.NewCache(),
		locks:           locks{keys: map[string]bool{}},
		reader:          source,
		cacheTTL:        ttl,
		expired:         expiredTime,
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

	if item.expired != time.UnixMicro(0) && !item.expired.After(time.Now()) {
		err = c.Delete(key)
		if err != nil {
			return nil, fmt.Errorf("error deleted item (expired time). err: %s", err)
		}
		return nil, ErrorItemExpired
	}

	if cachedValue, ok := item.cache.Get(cacheKeyPrefix + key); ok {
		return cachedValue, nil
	}

	// Если стоит блокировка, значит кто-то уже обновляет кеш. В этом случае
	// пытаемся отдать предыдущее значение.
	if item.locks.Get(key) {
		return c.tryToGetOldValue(key)
	}

	return c.updateCacheValue(key, item)
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

	item.cache.SetWithTTL(cacheKeyPrefix+key, result, item.cacheTTL)
	item.persistentCache.Set(cacheKeyPrefix+key, result)

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
		if cachedValue, ok := item.persistentCache.Get(cacheKeyPrefix + key); ok {
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
// ttl - время жизни записи в кеше, после удаляется с GC (0 - не удаляется никогда)
func CacheInit(ttl time.Duration) {
	d := cache{
		items: map[string]*cacheItem{},
		mx:    sync.RWMutex{},
		ttl:   ttl,
	}
	cacheCollection = d
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
