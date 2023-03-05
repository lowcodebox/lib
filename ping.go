package lib

import (
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

	service_port metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_port",
	}, []string{"value"})

	service_pid metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_pid",
	}, []string{"value"})

	service_replicas metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_replicas",
	}, []string{"value"})

	service_https metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_https",
	}, []string{"value"})

	service_dead_time metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_dead_time",
	}, []string{"value"})

	service_follower metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_follower",
	}, []string{"value"})

	service_grpc metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_grpc",
	}, []string{"value"})

	service_metric metrics.Gauge = kitprometheus.NewGaugeFrom(prometheus.GaugeOpts{
		Name: "service_metric",
	}, []string{"value"})
)

func SendServiceParamsToMetric(uid, name, version, status, port, pid, replicas, https, dead_time, follower, grpc, metric string, count float64) {
	service_uid.With("value", uid).Set(count)
	service_name.With("value", name).Set(count)
	service_version.With("value", version).Set(count)
	service_status.With("value", status).Set(count)
	service_port.With("value", port).Set(count)
	service_pid.With("value", pid).Set(count)
	service_replicas.With("value", replicas).Set(count)
	service_https.With("value", https).Set(count)
	service_dead_time.With("value", dead_time).Set(count)
	service_follower.With("value", follower).Set(count)
	service_grpc.With("value", grpc).Set(count)
	service_metric.With("value", metric).Set(count)
}
