package service

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"git.lowcodeplatform.net/fabric/models"
)


// Ping ...
func (s *service) Ping(ctx context.Context) (result []models.Pong, err error) {
	pp := strings.Split(s.cfg.Domain, "/")
	name := "ru"
	version := "ru"

	if len(pp) == 1 {
		name = pp[0]
	}
	if len(pp) == 2 {
		name = pp[0]
		version = pp[1]
	}

	pg, _ := strconv.Atoi(s.cfg.PortApp)
	pid := strconv.Itoa(os.Getpid())+":"+s.cfg.UidService
	state, _ := json.Marshal(s.metrics.Get())

	https := false
	if s.cfg.HttpsOnly != "" {
		https = true
	}
	var r = []models.Pong{
		{ s.cfg.DataUid,name, version, "run",pg, pid, string(state),s.cfg.ReplicasApp.Value, https, 0, ""},
	}

	return r, err
}