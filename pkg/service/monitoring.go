package service

import (
	"time"

	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

// Названия полей
var (
	fieldMethod = "method"
	fieldBlock  = "block"
)

// Метрики
var (
	timingServiceMetrics metrics.Histogram = kitprometheus.NewSummaryFrom(prometheus.SummaryOpts{
		Name:       "app_request_service_timing",
		Help:       "timing a request in service",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001},
	}, []string{fieldMethod})

	timingBlockMetrics metrics.Histogram = kitprometheus.NewSummaryFrom(prometheus.SummaryOpts{
		Name:       "app_request_block_timing",
		Help:       "timing a request in Blocks",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001},
	}, []string{fieldBlock})

	errorsMetrics metrics.Counter = kitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Name: "request_http_errors",
		Help: "error in http server",
	}, []string{fieldMethod})
)

func (*service) monitoringTimingService(method string, start time.Time) {
	timingServiceMetrics.With(fieldMethod, method).Observe(time.Since(start).Seconds())
}

func (*service) monitoringTimingBlock(block string, start time.Time) {
	timingBlockMetrics.With(fieldBlock, block).Observe(time.Since(start).Seconds())
}

func (*service) monitoringError(method string, err error) {
	if err == nil {
		return
	}

	errorsMetrics.With(fieldMethod, method).Add(1)
}
