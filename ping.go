package lib

import (
	"os"
	"runtime"
	"strings"
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
		if pingConfOld.Domain != "" && !strings.Contains(pingConfOld.Domain, "/") {
			elements := strings.Split(pingConfOld.Domain, "/")
			pingConf.Project, pingConf.Name = elements[0], elements[1]
		}

		if pingConf.PortHttp.Value == 0 && !pingConf.HttpsOnly.Value {
			pingConf.PortHttp = pingConf.Port
		}

		if pingConf.PortHttps.Value == 0 && pingConf.HttpsOnly.Value {
			pingConf.PortHttps = pingConf.Port
		}

		pongObj = models.PongObj{
			Uid:          FirstVal(pingConf.Uid, pingConfOld.DataUid),
			ReplicaID:    ksuid.New().String(),
			ProjectUid:   pingConf.Projectuid,
			Project:      FirstVal(pingConf.Project, pingConf.ProjectPointsrc),
			Name:         pingConf.Name,
			Service:      FirstVal(pingConf.Service, pingConfOld.ServiceType),
			Version:      pingConf.Version,
			HashCommit:   pingConf.HashCommit,
			Status:       "run",
			Pid:          os.Getpid(),
			Replicas:     FirstVal(pingConf.Replicas.Value, pingConfOld.ReplicasService.Value),
			PortHTTP:     pingConf.PortHttp.Value,
			PortHTTPS:    pingConf.PortHttps.Value,
			PortGrpc:     pingConf.PortGrpc.Value,
			Follower:     FirstVal(pingConf.Follower, pingConfOld.ServicePreloadPointsrc),
			Environment:  FirstVal(pingConf.Environment, pingConf.EnvironmentPointsrc),
			Cluster:      pingConf.Cluster,
			StartTime:    startTime,
			DC:           pingConf.DC,
			Mask:         pingConf.Mask,
			AccessPublic: pingConf.AccessPublic.Value,

			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		}
	}

	pongObj.Uptime = time.Since(pongObj.StartTime).String()

	return pongObj
}

func SetPongFields(f func(p *models.PongObj)) {
	if f == nil {
		return
	}

	if pongObj.ReplicaID == "" {
		Ping()
	}

	f(&pongObj)
}
