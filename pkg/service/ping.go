package service

import (
	"context"
	"os"
	"strconv"

	"git.lowcodeplatform.net/fabric/models"
	dto "github.com/prometheus/client_model/go"
)

// Ping ...
func (s *service) Ping(ctx context.Context) (result []models.Pong, err error) {
	var mobj []*dto.MetricFamily

	//metrics, err := cache.Cache().Get("prometheus")
	//if err != nil {
	//	metrics = fmt.Sprintf("error. %s", err)
	//	logger.Error(ctx, "cache.Cache", zap.Error(err))
	//}

	//bmetric, ok := metrics.([]byte)
	//if ok {
	//	err = json.Unmarshal(bmetric, &mobj)
	//	if err != nil {
	//		metrics = fmt.Sprintf("error. %s", err)
	//		logger.Error(ctx, "cache.Cache Unmarshal",
	//			zap.Error(err),
	//			zap.String("metrics failed body", fmt.Sprintf("%+v", metrics)))
	//	}
	//}

	pg, err := strconv.Atoi(s.cfg.PortApp)
	if err != nil {
		return nil, err
	}

	https := false

	pong := models.Pong{}
	pong.Uid = s.cfg.HashRun
	pong.Name = s.cfg.Name
	pong.Version = s.cfg.Version
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
	pong.Environment = s.cfg.Environment
	pong.AccessPublic = s.cfg.AccessPublic.Value

	pong.State = ""
	pong.Https = https

	result = append(result, pong)

	return result, err
}
