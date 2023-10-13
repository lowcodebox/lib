package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/ReneKroon/ttlcache"

	"github.com/pkg/errors"
)

const cacheKeyPrefix = "cache."

var (
	cacheCollection cache
)

type cache struct {
	items map[string]*cacheItem
	mx    sync.RWMutex
}

type cacheItem struct {
	// Getter определяет механизм получения данных от любого источника к/р поддерживает интерфейс
	reader          Reader
	cache           *ttlcache.Cache
	persistentCache *ttlcache.Cache
	locks           locks
	cacheTTL        time.Duration
}

type Reader interface {
	ReadSource() (res []byte, err error)
}

func Cache() *cache {
	if &cacheCollection == nil {
		panic("cache has not been initialized, call CacheRegister() before use")
	}

	return &cacheCollection
}

// Register регистрируем новый кеш (указываем фукнцию, кр будет возвращать нужное значение)
func (c *cache) Register(key string, source Reader, ttl time.Duration) (err error) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.items = map[string]*cacheItem{}

	cache := ttlcache.NewCache()
	cache.SkipTtlExtensionOnHit(true)

	ci := cacheItem{
		cache:           cache,
		persistentCache: ttlcache.NewCache(),
		locks:           locks{keys: map[string]bool{}},
		reader:          source,
		cacheTTL:        ttl,
	}
	c.items[key] = &ci
	return err
}

// Unregister
func (c *cache) Unregister(key string) (err error) {
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
		return nil, fmt.Errorf("error. key is not found")
	}

	if item.cache == nil {
		return nil, fmt.Errorf("cache is not inited")
	}

	if item.persistentCache == nil {
		return nil, fmt.Errorf("persistent cache is not inited")
	}

	if cachedValue, ok := item.cache.Get(cacheKeyPrefix + key); ok {
		return cachedValue, nil
	}

	// Если стоит блокировка, значит кто-то уже обновляет кеш. В этом случае
	// пытаемся отдать предыдущее значение.
	if item.locks.Get(key) {
		return c.tryToGetOldValue(key)
	}

	// Значение не найдено. Первый из запросов блокирует за собой обновление (на самом деле
	// может возникнуть ситуация когда несколько запросов поставят блокировку и начнут
	// обновлять кеш - пока считаем это некритичным).
	item.locks.Set(key, true)
	defer item.locks.Set(key, false)

	var values []byte
	values, err = item.reader.ReadSource()
	if err != nil {
		return nil, errors.Wrap(err, "could not get value from getter")
	}

	value = values

	item.cache.SetWithTTL(cacheKeyPrefix+key, value, item.cacheTTL)
	item.persistentCache.Set(cacheKeyPrefix+key, value)

	return value, nil
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
// source - источник, откуда мы получаем значения для кеширования
func CacheRegister() {
	d := cache{
		items: map[string]*cacheItem{},
		mx:    sync.RWMutex{},
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
