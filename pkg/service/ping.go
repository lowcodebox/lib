package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/cache"

	"git.lowcodeplatform.net/packages/logger"
	dto "github.com/prometheus/client_model/go"

	"go.uber.org/zap"
)

// Ping ...
func (s *service) Ping(ctx context.Context) (result []models.Pong, err error) {
	var mobj []*dto.MetricFamily

	metrics, err := cache.Cache().Get("prometheus")
	if err != nil {
		metrics = fmt.Sprintf("error. %s", err)
		logger.Error(ctx, "cache.Cache", zap.Error(err))
	}

	err = json.Unmarshal(metrics.([]byte), &mobj)
	if err != nil {
		metrics = fmt.Sprintf("error. %s", err)
		logger.Error(ctx, "cache.Cache Unmarshal", zap.Error(err))
	}

	pg, _ := strconv.Atoi(s.cfg.PortApp)

	https := false

	version := s.cfg.ServiceType
	splDomain := strings.Split(s.cfg.Domain, "/")
	if len(splDomain) == 2 {
		version = splDomain[1]
	}

	pong := models.Pong{}
	pong.Uid = s.cfg.DataUid
	pong.Name = s.cfg.Name
	pong.Version = version
	pong.Status = "run"
	pong.PortHTTP, pong.Port = pg, pg
	pong.Config = s.cfg.ConfigName
	pong.Pid = strconv.Itoa(os.Getpid())
	pong.Replicas = s.cfg.Replicas.Value
	pong.EnableHttps = false
	pong.PortGrpc = 0
	pong.PortMetric = 8080
	pong.Metrics = mobj
	pong.ServiceVersion = s.cfg.ServiceVersion
	pong.HashCommit = s.cfg.HashCommit

	pong.State = ""
	pong.Https = https

	result = append(result, pong)

	return result, err
}