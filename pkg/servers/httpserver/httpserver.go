package httpserver

import (
	"context"
	"fmt"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/app/pkg/service"
	"git.lowcodeplatform.net/fabric/app/pkg/session"
	iam "git.lowcodeplatform.net/fabric/iam-client"
	"git.lowcodeplatform.net/fabric/lib"
	bbmetric "git.lowcodeplatform.net/fabric/lib"
	"github.com/labstack/gommon/color"
	"net/http"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	// should be so!
	_ "git.lowcodeplatform.net/fabric/app/pkg/servers/docs"
)

type httpserver struct {
	ctx     context.Context
	cfg     model.Config
	src     service.Service
	metric  bbmetric.ServiceMetric
	logger  lib.Log
	iam     iam.IAM
	session session.Session
}

type Server interface {
	Run() (err error)
}

// Run server
func (h *httpserver) Run() error {
	done := color.Green("[OK]")

	// закрываем логи при завешрении работы сервера
	defer func() {
		h.logger.Warning("Service is stopped. Logfile is closed.")
		h.logger.Close()
	}()

	//err := httpscerts.Check(h.cfg.SSLCertPath, h.cfg.SSLPrivateKeyPath)
	//if err != nil {
	//	panic(err)
	//}
	srv := &http.Server{
		Addr:         ":" + h.cfg.PortApp,
		Handler:      h.NewRouter(false),	// переадресация будет работать, если сам севрис будет стартовать https-сервер (для этого надо получать сертфикаты)
		ReadTimeout:  h.cfg.ReadTimeout.Value,
		WriteTimeout: h.cfg.WriteTimeout.Value,
	}
	fmt.Printf("%s Service run (port:%s)\n", done, h.cfg.PortApp)
	h.logger.Info("Запуск https сервера", zap.String("port", h.cfg.PortApp))
	//e := srv.ListenAndServeTLS(h.cfg.SSLCertPath, h.cfg.SSLPrivateKeyPath)

	e := srv.ListenAndServe()
	if e != nil {
		return errors.Wrap(e, "SERVER run")
	}
	return nil
}


func New(
	ctx 	context.Context,
	cfg 	model.Config,
	src 	service.Service,
	metric 	bbmetric.ServiceMetric,
	logger 	lib.Log,
	iam 	iam.IAM,
	session session.Session,
) Server {
	return &httpserver{
		ctx,
		cfg,
		src,
		metric,
		logger,
		iam,
		session,
	}
}