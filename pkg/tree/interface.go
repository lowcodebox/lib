package tree

import (
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
)

type tree struct {
	cfg model.Config
}

type Tree interface {
	DataToIncl(objData []models.Data) []*models.DataTree
	TreeShowIncl(in []*models.DataTree, obj string) (out []*models.DataTree)
	SortItems(p []*models.DataTree, fieldsort string, typesort string)
}

func New(cfg model.Config) Tree {
	return &tree{
		cfg,
	}
}
