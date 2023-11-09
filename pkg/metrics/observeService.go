package metrics

import (
	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	txMetrics metrics.Histogram = kitprometheus.NewSummaryFrom(prometheus.SummaryOpts{
		Name:       "request_exec_service_timing_app",
		Help:       "timing a request to the app methods",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001},
	}, []string{"service_id", "method"})
)
