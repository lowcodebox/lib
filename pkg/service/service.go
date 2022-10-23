package service

import (
	"context"

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
)

type service struct {
	logger   lib.Log
	cfg      model.Config
	metrics  lib.ServiceMetric
	cache    cache.Cache
	block    block.Block
	function function.Function
	msg      i18n.I18n
	session  session.Session
	api      api.Api
	iam      iam.IAM
	vfs      lib.Vfs
}

// Service interface
type Service interface {
	Alive(ctx context.Context) (out model.AliveOut, err error)
	Storage(ctx context.Context, in model.StorageIn) (out model.StorageOut, err error)
	Ping(ctx context.Context) (result []models.Pong, err error)
	Page(ctx context.Context, in model.ServiceIn) (out model.ServicePageOut, err error)
	Block(ctx context.Context, in model.ServiceIn) (out model.ServiceBlockOut, err error)
	Cache(ctx context.Context, in model.ServiceCacheIn) (out model.RestStatus, err error)
	AuthChangeRole(ctx context.Context, in model.ServiceAuthIn) (out model.ServiceAuthOut, err error)
}

func New(
	logger lib.Log,
	cfg model.Config,
	metrics lib.ServiceMetric,
	cache cache.Cache,
	msg i18n.I18n,
	session session.Session,
	api api.Api,
	iam iam.IAM,
	vfs lib.Vfs,
) Service {
	var tplfunc = function.NewTplFunc(cfg, logger, api)
	var function = function.New(cfg, logger, api)
	var blocks = block.New(cfg, logger, function, tplfunc, api)

	return &service{
		logger,
		cfg,
		metrics,
		cache,
		blocks,
		function,
		msg,
		session,
		api,
		iam,
		vfs,
	}
}
