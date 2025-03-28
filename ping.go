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
	configName  string
	startTime   = time.Now()
)

func Ping() models.PongObj {
	if pongObj.ReplicaID != "" {
		pongObj.Uptime = time.Since(pongObj.StartTime).String()

		return pongObj
	}

	if pingConfOld.Domain != "" && strings.Contains(pingConfOld.Domain, "/") {
		elements := strings.Split(pingConfOld.Domain, "/")
		pingConf.Project, pingConf.Name = elements[0], elements[1]
	}

	if pingConf.PortHttp == 0 && !pingConf.HttpsOnly.V() {
		pingConf.PortHttp = pingConf.Port
	}

	if pingConf.PortHttps.V() == 0 && pingConf.HttpsOnly.V() {
		pingConf.PortHttps = pingConf.Port
	}

	pongObj = models.PongObj{
		Uid:          FirstVal(pingConf.Uid, pingConfOld.DataUid, configName),
		ReplicaID:    ksuid.New().String(),
		ProjectUid:   pingConf.Projectuid,
		Project:      FirstVal(pingConf.Project, pingConf.ProjectPointsrc),
		Name:         FirstVal(pingConf.Name, pingConf.Service, pingConfOld.ServiceType),
		Service:      FirstVal(pingConf.Service, pingConfOld.ServiceType),
		Version:      pingConf.Version,
		HashCommit:   pingConf.HashCommit,
		Status:       "run",
		Pid:          os.Getpid(),
		Replicas:     FirstVal(pingConf.Replicas.V(), pingConfOld.ReplicasService.V()),
		PortHTTP:     pingConf.PortHttp.V(),
		PortHTTPS:    pingConf.PortHttps.V(),
		PortGrpc:     pingConf.PortGrpc.V(),
		Follower:     FirstVal(pingConf.Follower, pingConfOld.ServicePreloadPointsrc),
		Environment:  FirstVal(pingConf.Environment, pingConf.EnvironmentPointsrc),
		Cluster:      pingConf.Cluster,
		StartTime:    startTime,
		Uptime:       time.Since(startTime).String(),
		DC:           pingConf.DC,
		Mask:         pingConf.Mask,
		AccessPublic: pingConf.AccessPublic.V(),

		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

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
