package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"git.edtech.vm.prod-6.cloud.el/fabric/api-client"
	iam "git.edtech.vm.prod-6.cloud.el/fabric/iam-client"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"github.com/labstack/gommon/color"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/service"
	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/session"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	applib "git.edtech.vm.prod-6.cloud.el/fabric/app/lib"

	// should be so!
	_ "git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/servers/docs"
)

type httpserver struct {
	ctx     context.Context
	cfg     model.Config
	src     service.Service
	iam     iam.IAM
	session session.Session
	vfs     lib.Vfs
	api     api.Api
	app_lib applib.App

	serviceVersion string
	hashCommit     string
}

type Server interface {
	Run() (err error)
}

// Run server
func (h *httpserver) Run() error {
	done := color.Green("[OK]")
	fail := color.Red("[NO]")

	//err := httpscerts.Check(h.cfg.SSLCertPath, h.cfg.SSLPrivateKeyPath)
	//if err != nil {
	//	panic(err)
	//}
	router, err := h.NewRouter(false)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr:         ":" + h.cfg.PortApp,
		Handler:      router, // переадресация будет работать, если сам севрис будет стартовать https-сервер (для этого надо получать сертфикаты)
		ReadTimeout:  h.cfg.ReadTimeout.Value,
		WriteTimeout: h.cfg.WriteTimeout.Value,
	}

	securityModeDesc := ""
	securityMode := false

	if h.cfg.Signin == "checked" && h.cfg.SigninUrl != "" {
		securityModeDesc = " (security mode: enable)"
		securityMode = true
	}

	fmt.Printf("%s Service run%s (port:%s)\n", done, securityModeDesc, h.cfg.PortApp)
	logger.Info(h.ctx, "Запуск https сервера", zap.Bool("security mode", securityMode), zap.String("port", h.cfg.PortApp))
	//e := srv.ListenAndServeTLS(h.cfg.SSLCertPath, h.cfg.SSLPrivateKeyPath)

	e := srv.ListenAndServe()
	if e != nil {
		fmt.Printf("%s Error run (port:%s) err: %s\n", fail, h.cfg.PortApp, e)
		return errors.Wrap(e, "SERVER run")
	}
	return nil
}

func New(
	ctx context.Context,
	cfg model.Config,
	src service.Service,
	iam iam.IAM,
	session session.Session,
	vfs lib.Vfs,
	api api.Api,
	app_lib applib.App,
	serviceVersion string,
	hashCommit string,
) Server {
	return &httpserver{
		ctx,
		cfg,
		src,
		iam,
		session,
		vfs,
		api,
		app_lib,
		serviceVersion,
		hashCommit,
	}
}
