package tree

import (
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
)

type tree struct {
	cfg    	model.Config
	logger 	lib.Log
}

type Tree interface {
	DataToIncl(objData []models.Data) []*models.DataTree
	TreeShowIncl(in []*models.DataTree, obj string) (out []*models.DataTree)
	SortItems(p []*models.DataTree, fieldsort string, typesort string)
}

func New(cfg model.Config, logger lib.Log) Tree {
	return &tree{
		cfg,
		logger,
	}
}