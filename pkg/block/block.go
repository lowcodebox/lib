package block

import (
	"context"
	"path/filepath"
	"sync"

	"git.lowcodeplatform.net/fabric/api-client"
	"git.lowcodeplatform.net/fabric/app/pkg/cache"
	"git.lowcodeplatform.net/fabric/app/pkg/function"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
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
	GetWithLocalCache(ctx context.Context, in model.ServiceIn, block, page models.Data, values map[string]interface{}) (moduleResult model.ModuleResult, err error)
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
