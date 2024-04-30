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

## v0.1.16 dev
Переписан пакет с использованием sync.Map для избежания паник

Результаты сравнения тестов Map и sync.Map в контексте данного пакета

## Результаты тестирования производительности

В таблице ниже представлены результаты тестирования производительности с использованием `Map` и `sync.Map`:

| Тест                                      | Время (ns/op) с Map | Время (ns/op) с sync.Map | Время, быстрее в раз | B/op с Map | B/op с sync.Map | Сокращение B/op | allocs/op с Map | allocs/op с sync.Map | Сокращение allocs/op |
|------------------------------------------|---------------------|--------------------------|----------------------|------------|-----------------|------------------|-----------------|----------------------|-----------------------|
| BenchmarkCache_ConcurrentUpsertAndGet    | 8,298,061,621       | 509,060,644              | 16.3                 | 400,676,440| 212,454,344     | 47%              | 4,802,734       | 2,904,396            | 39.5%                 |
| BenchmarkCache_ConcurrentUpsert          | 80,735              | 10,448                   | 7.7                  | 3,796      | 4,172           | -9.9%            | 43              | 54                    | -25.6%                |
| BenchmarkCache_ConcurrentGet2            | 607.5               | 605.9                    | 1.002                | 15         | 15              | 0%               | 1               | 1                     | 0%                    |
| BenchmarkCache_ProductionLike            | 94,983              | 63,739                   | 1.49                 | 917        | 1,001           | -9.2%            | 25              | 33                    | -32%                  |

Использование `sync.Map` приводит к значительному увеличению производительности по сравнению с использованием обычного `Map` в большинстве тестов. Особенно это заметно в тесте `BenchmarkCache_ConcurrentUpsertAndGet`, где время выполнения уменьшилось в 16.3 раза. Наблюдается сокращение использования памяти (B/op) и количество аллокаций (allocs/op) в некоторых тестах, хотя в одном случае количество аллокаций увеличилось.
