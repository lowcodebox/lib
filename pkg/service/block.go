package service

import (
	"context"
	"fmt"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/models"
)

func (s *service) Block(ctx context.Context, in model.ServiceIn) (out model.ServiceBlockOut, err error) {
	start := time.Now()
	defer s.monitoringTimingService("Block", start)
	defer s.monitoringError("Block", err)

	var objBlock *models.ResponseData

	//t1 := time.Now()
	//rnd := lib.UUID()

	dataPage := models.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы
	objBlock, err = s.api.ObjGetWithCache(ctx, in.Block)
	//logger.Info(ctx, "gen block",
	//	zap.String("block", in.Block),
	//	zap.String("step", "получение блока (ObjGetWithCache)"),
	//	zap.Float64("timing", time.Since(t1).Seconds()),
	//	zap.String("rnd", rnd))

	if err != nil {
		return out, fmt.Errorf("error get obj with cache. block: %s, err: %s", in.Block, err)
	}
	if objBlock == nil {
		return out, fmt.Errorf("error. lenght data from objBlock is 0. block: %s", in.Block)
	}
	if len(objBlock.Data) == 0 {
		return out, fmt.Errorf("error. lenght data from objBlock.Data is 0. block: %s", in.Block)
	}
	//t2 := time.Now()

	//logger.Info(ctx, "gen block",
	//	zap.Any("objBlock.Data[0]", objBlock.Data[0]),
	//	zap.String("step", " блока (полная)"),
	//	zap.Float64("timing", time.Since(t2).Seconds()), zap.String("rnd", rnd))

	moduleResult, err := s.block.Get(ctx, in, objBlock.Data[0], dataPage, nil)
	if err != nil {
		return out, fmt.Errorf("error get obj block: %s, err: %s", in.Block, err)
	}

	//logger.Info(ctx, "gen block",
	//	zap.String("block", in.Block),
	//	zap.String("step", "генерация блока (полная)"),
	//	zap.Float64("timing", time.Since(t2).Seconds()), zap.String("rnd", rnd))

	out.Result = moduleResult.Result

	s.monitoringTimingBlock(moduleResult.Id, start)
	return
}
