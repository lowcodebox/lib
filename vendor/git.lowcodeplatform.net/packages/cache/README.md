# cache

Для работы с кешем, его надо инициировать

```Init(ctx context.Context, expiredInterval, runGCInterval time.Duration)```

где:
- expiredInterval - время жизни записи в кеше, после удаляется с GC (0 - не удаляется никогда)
- runGCInterval - интервал срабатывания GC

Далее создаем/обновляем записи кеша

```Upsert(key string, source func() (res interface{}, err error), refreshInterval time.Duration)```

Пример:

```		
err = cache.Cache().Upsert(key, func() (res interface{}, err error) {
    res, err = a.Query(ctx, query, method, bodyJSON)
    return res, err
}, a.updateTime)

```

Получаем значение кеша

``Get(key string) (value interface{}, err error)``

```
value, err = cache.Cache().Get(key)
if err != nil {
    err = fmt.Errorf("get value is fail. err: %s", err)
}
```