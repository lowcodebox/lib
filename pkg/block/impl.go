package block

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"sync"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

// Get получение содержимого блока (с учетом операций с кешем)
func (s *block) Get(ctx context.Context, in model.ServiceIn, block, page models.Data, values map[string]interface{}) (moduleResult model.ModuleResult, err error) {

	var addСonditionPath bool
	var addСonditionURL bool
	var cacheInterval int

	cacheInt, _ := block.Attr("cache", "value") // включен ли режим кеширования
	cache_nokey2, _ := block.Attr("cache_keyAddPath", "value")
	cache_nokey3, _ := block.Attr("cache_keyAddURL", "value")

	if cache_nokey2 == "checked" {
		addСonditionPath = true
	}
	if cache_nokey3 == "checked" {
		addСonditionURL = true
	}

	t1 := time.Now()

	// если интервал не задан, то не кешируем
	if cacheInt == "" {
		cacheInt = "0"
	}
	cacheInterval, err = strconv.Atoi(cacheInt)
	if err != nil {
		cacheInterval = 0
		err = nil
	}

	// если включен кеш и есть интервал кеширования
	if s.cache.Active() && cacheInterval != 0 {

		// читаем из кеша и отдаем (ВСЕГДА сразу)
		key, cacheParams := s.cache.GenKey(block.Uid, in.CachePath, in.CacheQuery, addСonditionPath, addСonditionURL)
		result, _, flagExpired, err := s.cache.Read(key)

		// 1 кеша нет (срабатывает только при первом формировании)
		if err != nil {
			logger.Error(ctx, "err get cache (GetBlock)", zap.String("step", "err get cache"),
				zap.Float64("timing", time.Since(t1).Seconds()),
				zap.String("result", result), zap.String("block.Id", block.Id), zap.String("key", key), zap.Error(err))

			result, err = s.updateCache(ctx, key, cacheParams, cacheInterval, in, block, page, values)
			if err != nil {
				logger.Info(ctx, "err update cache (GetBlock)", zap.String("step", "err update cache"),
					zap.Float64("timing", time.Since(t1).Seconds()),
					zap.String("result", result), zap.String("block.Id", block.Id), zap.String("key", key), zap.Error(err))
			}
		} else {
			// 2 время закончилось (не обращаем внимание на статус "обновляется" потому, что при изменении статуса на "обновляем"
			// мы увеличиваем время на предельно время проведения обновления
			// требуется обновить фоном (отдали текущие данные из кеша)
			if flagExpired {
				go s.updateCache(ctx, key, cacheParams, cacheInterval, in, block, page, values)
			}
		}

		moduleResult = model.ModuleResult{
			Id:     block.Id,
			Result: template.HTML(result),
			Stat:   nil,
			Err:    nil,
		}

	} else {
		mResult, err := s.generate(ctx, in, block, page, values)
		if err != nil {
			moduleResult.Result = ""
			moduleResult.Err = err
			return moduleResult, err
		}

		moduleResult = mResult
	}

	return
}

// GetToChannel получаем содержимое блока в передачей через канал
func (b *block) GetToChannel(ctx context.Context, in model.ServiceIn, block, page models.Data, values map[string]interface{}, buildChan chan model.ModuleResult, wg *sync.WaitGroup) (err error) {
	defer wg.Done()

	// проверка на выход по сигналу
	select {
	case <-ctx.Done():
		return
	default:
	}

	moduleResult, err := b.Get(ctx, in, block, page, values)
	if err != nil {
		moduleResult.Err = err
		moduleResult.Result = template.HTML(fmt.Sprint(err))
	}
	buildChan <- moduleResult

	return
}