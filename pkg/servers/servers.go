// запускаем указанные виды из поддерживаемых серверов
package servers

import (
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/app/pkg/servers/httpserver"
	"git.lowcodeplatform.net/fabric/app/pkg/service"
	bbmetric "git.lowcodeplatform.net/fabric/lib"
	"strings"
)

type servers struct {
	mode       string
	service    service.Service
	httpserver httpserver.Server
	metrics    bbmetric.ServiceMetric
	cfg        model.Config
}

type Servers interface {
	Run()
}

// запускаем указанные севрера
func (s *servers) Run() {
	if strings.Contains(s.mode, "http") {
		s.httpserver.Run()
	}
}

func New(
	mode string,
	service service.Service,
	httpserver httpserver.Server,
	metrics bbmetric.ServiceMetric,
	cfg model.Config,
) Servers {
	return &servers{
		mode,
		service,
		httpserver,
		metrics,
		cfg,
	}
}