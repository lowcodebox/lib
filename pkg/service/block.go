package service

import (
	"context"
	"fmt"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/models"
)

func (s *service) Block(ctx context.Context, in model.ServiceIn) (out model.ServiceBlockOut, err error) {
	var objBlock *models.ResponseData

	dataPage := models.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы
	objBlock, err = s.api.ObjGetWithCache(ctx, in.Block)
	if err != nil {
		return out, fmt.Errorf("error get obj with cache. block: %s, err: %s", in.Block, err)
	}
	if objBlock == nil {
		return out, fmt.Errorf("error. lenght data from objBlock is 0. block: %s", in.Block)
	}
	if len(objBlock.Data) == 0 {
		return out, fmt.Errorf("error. lenght data from objBlock is 0. block: %s", in.Block)
	}

	moduleResult, err := s.block.Get(ctx, in, objBlock.Data[0], dataPage, nil)
	if err != nil {
		return out, fmt.Errorf("error get obj block: %s, err: %s", in.Block, err)
	}

	out.Result = moduleResult.Result

	return
}
