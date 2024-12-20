package lib

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// uid
	service_uid metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_uid",
	}, []string{"value"})

	service_name metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_name",
	}, []string{"value"})

	service_version metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_version",
	}, []string{"value"})

	service_status metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_status",
	}, []string{"value"})

	service_port_http metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_port_http",
	}, []string{"value"})

	service_pid metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_pid",
	}, []string{"value"})

	service_replicas metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_replicas",
	}, []string{"value"})

	service_port_https metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_port_https",
	}, []string{"value"})

	service_dead_time metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_dead_time",
	}, []string{"value"})

	service_follower metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_follower",
	}, []string{"value"})

	service_port_grpc metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_port_grpc",
	}, []string{"value"})

	service_port_metrics metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_port_metrics",
	}, []string{"value"})

	buildInfo *prometheus.GaugeVec
)

func SendServiceParamsToMetric(uid, name, version, status, pid, replicas, portHTTP, portGRPC, portMetrics, portHTTPS, dead_time, follower string) {
	var count float64
	service_uid.With("value", uid).Set(count)
	service_name.With("value", name).Set(count)
	service_version.With("value", version).Set(count)
	service_status.With("value", status).Set(count)
	service_port_http.With("value", portHTTP).Set(count)
	service_pid.With("value", pid).Set(count)
	service_replicas.With("value", replicas).Set(count)
	service_port_https.With("value", portHTTPS).Set(count)
	service_dead_time.With("value", dead_time).Set(count)
	service_follower.With("value", follower).Set(count)
	service_port_grpc.With("value", portGRPC).Set(count)
	service_port_metrics.With("value", portMetrics).Set(count)
}

// ValidateNameVersion - формирует правильные имя проекта и версию сервиса исходя из того, что пришло из настроек
func ValidateNameVersion(project, types, domain string) (name, version string) {
	name = "unknown"

	if project != "" {
		if len(strings.Split(project, "-")) > 3 { // признак того, что получили UID (для совместимости)
			if domain != "" {
				project = strings.Split(domain, "/")[0]
			}
		}
		name = project // название проекта
	}

	pp := strings.Split(domain, "/")
	if len(pp) == 1 {
		if pp[0] != "" {
			name = pp[0]
		}
	}
	if len(pp) == 2 {
		if pp[0] != "" {
			name = pp[0]
		}
		if pp[1] != "" {
			version = pp[1]
		}
	}

	if version == "" {
		version = types
	}

	return name, version
}

func NewBuildInfo(service string) *prometheus.GaugeVec {
	buildInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: service,
		Name:      "build_info",
		Help: fmt.Sprintf(
			"A metric with a constant '1' value labeled by version, revision, branch, and goversion from which %s was built.",
			service,
		),
	}, []string{"version", "revision", "dc", "goversion"})

	return buildInfo
}

func SetBuildInfo(version, revision, dc string) {
	buildInfo.WithLabelValues(
		version, revision, dc, runtime.Version(),
	).Set(1)
}
