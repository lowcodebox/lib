package block

import (
	"context"
	"path/filepath"
	"sync"

	api "git.edtech.vm.prod-6.cloud.el/fabric/api-client"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/cache"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/function"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
)

const sep = string(filepath.Separator)
const prefixUploadURL = "upload" // адрес/_prefixUploadURL_/... - путь, относительно bucket-а проекта

type block struct {
	cfg      model.Config
	function function.Function
	tplfunc  function.TplFunc
	api      api.Api
	vfs      lib.Vfs
	cache    cache.Cache
}

type Block interface {
	Get(ctx context.Context, in model.ServiceIn, block, page models.Data, values map[string]interface{}) (moduleResult model.ModuleResult, err error)
	GetToChannel(ctx context.Context, in model.ServiceIn, block, page models.Data, values map[string]interface{}, buildChan chan model.ModuleResult, wg *sync.WaitGroup) (err error)
}

func New(
	cfg model.Config,
	function function.Function,
	tplfunc function.TplFunc,
	api api.Api,
	vfs lib.Vfs,
	cache cache.Cache,
) Block {
	return &block{
		cfg:      cfg,
		function: function,
		tplfunc:  tplfunc,
		api:      api,
		vfs:      vfs,
		cache:    cache,
	}
}
