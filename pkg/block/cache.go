package block

import (
	"context"
	"fmt"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"go.uber.org/zap"
)

// updateCache внутренняя фунция сервиса.
// не вынесена в пакет Cache потому-что требуется генерировать блок
func (b *block) updateCache(ctx context.Context, key, cacheParams string, cacheInterval int, in model.ServiceIn, block models.Data, page models.Data, values map[string]interface{}) (result string, err error) {
	t1 := time.Now()
	logger.Info(ctx, "UpdateCache", zap.String("step", "start update cache"),
		zap.String("cacheParams", cacheParams),
		zap.String("result", result), zap.String("block.Id", block.Id), zap.String("key", key), zap.Error(err))

	err = b.cache.SetStatus(key, "updated")
	if err != nil {
		result = fmt.Sprint(err)

		logger.Error(ctx, "UpdateCache", zap.String("step", "err set status"),
			zap.String("cacheParams", cacheParams),
			zap.String("block.Id", block.Id), zap.String("key", key), zap.Error(err))
	}

	moduleResult, err := b.generate(ctx, in, block, page, values)
	if err != nil {
		result = fmt.Sprintf("Error [Generate] in updateCache from %s. Cache not saved. Time generate: %s. Error: %s", block.Id, time.Since(t1), err)

		logger.Error(ctx, "UpdateCache", zap.String("step", "err generate block"),
			zap.String("desc", result),
			zap.Float64("timing", time.Since(t1).Seconds()),
			zap.String("block.Id", block.Id), zap.String("key", key), zap.Error(err))

		return result, err
	}

	err = b.cache.Write(key, cacheParams, cacheInterval, block.Uid, page.Uid, string(moduleResult.Result))
	if err != nil {

		logger.Error(ctx, "UpdateCache", zap.String("step", "err write cache"),
			zap.Float64("timing", time.Since(t1).Seconds()),
			zap.String("block.Id", block.Id),
			zap.String("page.Uid", page.Uid),
			zap.String("key", key), zap.Error(err))
	}

	result = string(moduleResult.Result)

	logger.Info(ctx, "UpdateCache", zap.String("step", "finished update cache"),
		zap.String("cacheParams", cacheParams),
		zap.Int("result len", len(result)),
		zap.String("block.Id", block.Id), zap.String("key", key), zap.Error(err))

	return
}
