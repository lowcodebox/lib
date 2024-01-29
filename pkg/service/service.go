package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	api "git.lowcodeplatform.net/fabric/api-client"
	"git.lowcodeplatform.net/fabric/app/pkg/block"
	"git.lowcodeplatform.net/fabric/app/pkg/cache"
	"git.lowcodeplatform.net/fabric/app/pkg/function"
	"git.lowcodeplatform.net/fabric/app/pkg/i18n"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/app/pkg/session"
	iam "git.lowcodeplatform.net/fabric/iam-client"
	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

const queryPublicPages = "sys_public_pages"

// список роутеров, для который пропускается авторизация клиента
var constPublicLink = map[string]bool{
	"/ping":      true,
	"/templates": true,
	"/upload":    true,
	"/logout":    true,
}

// динамические параметры, которые могут меняться через асинхронные шедулеры (повышение производительности)
type dynamicParams struct {
	PublicPages  map[string]bool
	PublicRoutes map[string]bool
}

type service struct {
	ctx      context.Context
	cfg      model.Config
	cache    cache.Cache
	block    block.Block
	function function.Function
	msg      i18n.I18n
	session  session.Session
	api      api.Api
	iam      iam.IAM
	vfs      lib.Vfs
	dps      *dynamicParams
}

// Service interface
type Service interface {
	Alive(ctx context.Context) (out model.AliveOut, err error)
	Storage(ctx context.Context, in model.StorageIn) (out model.StorageOut, err error)
	Ping(ctx context.Context) (result []models.Pong, err error)
	Files(ctx context.Context, in model.ServiceFilesIn) (out model.ServiceFilesOut, err error)
	Page(ctx context.Context, in model.ServiceIn) (out model.ServicePageOut, err error)
	Block(ctx context.Context, in model.ServiceIn) (out model.ServiceBlockOut, err error)
	Cache(ctx context.Context, in model.ServiceCacheIn) (out model.RestStatus, err error)
	AuthChangeRole(ctx context.Context, in model.ServiceAuthChangeIn) (out model.ServiceAuthChangeOut, err error)
	AuthLogIn(ctx context.Context, in model.ServiceAuthIn) (out model.ServiceAuthOut, err error)

	GetDynamicParams() *dynamicParams
}

func New(
	ctx context.Context,
	cfg model.Config,
	cache cache.Cache,
	msg i18n.I18n,
	session session.Session,
	api api.Api,
	iam iam.IAM,
	vfs lib.Vfs,
) Service {
	var dps = dynamicParams{
		PublicPages:  map[string]bool{},
		PublicRoutes: constPublicLink,
	}

	var tplfunc = function.NewTplFunc(cfg, api)
	var function = function.New(cfg, api)
	var blocks = block.New(cfg, function, tplfunc, api, vfs, cache)

	// асинхронно обновляем список публичный страниц/блоков
	go reloadPublicPages(ctx, &dps, api, 10*time.Second)

	return &service{
		ctx,
		cfg,
		cache,
		blocks,
		function,
		msg,
		session,
		api,
		iam,
		vfs,
		&dps,
	}
}

// ReloadFromPG обновляем meta если обновилось время в кипере (изменили данные и нажали - обновить в сервисе)
func reloadPublicPages(ctx context.Context, d *dynamicParams, api api.Api, intervalReload time.Duration) {
	var objs = models.ResponseData{}
	ticker := time.NewTicker(intervalReload)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			res, err := api.Query(ctx, queryPublicPages, http.MethodGet, "")
			if err != nil {
				logger.Error(ctx, "error api.Query", zap.String("query", queryPublicPages), zap.Error(err))
				ticker = time.NewTicker(intervalReload)
				continue
			}
			err = json.Unmarshal([]byte(res), &objs)
			if err != nil {
				logger.Error(ctx, "error Unmarshal api.Query", zap.String("query", queryPublicPages), zap.Error(err))
				ticker = time.NewTicker(intervalReload)
				continue
			}

			resD := map[string]bool{}
			for _, v := range objs.Data {
				resD[v.Uid] = true
				if v.Id != "" {
					resD[v.Id] = true
				}
			}
			d.PublicPages = resD

			ticker = time.NewTicker(intervalReload)
		}
	}
}

func (s *service) GetDynamicParams() *dynamicParams {
	return s.dps
}
