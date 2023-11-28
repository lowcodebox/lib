package service

import (
	"context"
	"fmt"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/models"
)

func (s *service) Block(ctx context.Context, in model.ServiceIn) (out model.ServiceBlockOut, err error) {
	var objBlock models.ResponseData
	dataPage := models.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы

	objBlock, err = s.api.ObjGet(ctx, in.Block)

	if len(objBlock.Data) == 0 {
		return out, fmt.Errorf("%s", "Error. Lenght data from objBlock is 0.")
	}
	moduleResult, err := s.block.Generate(ctx, in, objBlock.Data[0], dataPage, nil)
	out.Result = moduleResult.Result

	return
}
