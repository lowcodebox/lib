package service

import (
	"context"
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
	pid := strconv.Itoa(os.Getpid()) + ":" + s.cfg.UidService
	state := ""

	https := false
	if s.cfg.HttpsOnly != "" {
		https = true
	}
	var r = []models.Pong{
		{
			Uid:      s.cfg.DataUid,
			Name:     name,
			Version:  version,
			Status:   "run",
			Port:     pg,
			Pid:      pid,
			State:    state,
			Replicas: s.cfg.ReplicasApp.Value,
			Https:    https,
			DeadTime: 0,
			Follower: ""},
	}

	return r, err
}
