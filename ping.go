package lib

import (
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"github.com/segmentio/ksuid"
)

var (
	pongObj     models.PongObj
	pingConf    models.PingConfig
	pingConfOld models.PingConfigOld
	startTime   = time.Now()
)

func Ping() models.PongObj {
	if pongObj.ReplicaID == "" {
		pongObj = models.PongObj{
			Uid:          pingConf.Uid,
			ReplicaID:    ksuid.New().String(),
			ProjectUid:   pingConf.Projectuid,
			Project:      pingConf.Project,
			Name:         pingConf.Name,
			Service:      pingConf.Service,
			Version:      pingConf.Version,
			HashCommit:   _,
			Status:       _,
			Host:         _,
			Pid:          _,
			Replicas:     _,
			PortHTTP:     _,
			PortHTTPS:    _,
			PortGrpc:     _,
			EnableHttps:  _,
			Follower:     _,
			Environment:  _,
			Cluster:      _,
			DeadTime:     _,
			StartTime:    startTime,
			DC:           _,
			Mask:         _,
			AccessPublic: _,
			OS:           _,
			Arch:         _,
			Code:         _,
			Error:        _,
		}
	}

	pongObj.Uptime = time.Since(pongObj.StartTime).String()

	return pongObj
}
