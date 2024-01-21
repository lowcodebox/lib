// запускаем указанные виды из поддерживаемых серверов
package servers

import (
	"strings"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/app/pkg/servers/httpserver"
	"git.lowcodeplatform.net/fabric/app/pkg/service"
)

type servers struct {
	mode       string
	service    service.Service
	httpserver httpserver.Server
	cfg        model.Config
}

type Servers interface {
	Run() error
}

// Run запускаем указанные севрера
func (s *servers) Run() (err error) {
	if strings.Contains(s.mode, "http") {
		err = s.httpserver.Run()
	}

	return err
}

func New(
	mode string,
	service service.Service,
	httpserver httpserver.Server,
	cfg model.Config,
) Servers {
	return &servers{
		mode,
		service,
		httpserver,
		cfg,
	}
}
