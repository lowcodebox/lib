package service

import (
	"time"

	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

// Названия полей
const (
	fieldMethod = "method"
	fieldBlock  = "block"
)

// Метрики
var (
	timingServiceMetrics metrics.Histogram = kitprometheus.NewHistogramFrom(prometheus.HistogramOpts{
		Name:    "app_request_service_timing",
		Help:    "timing a request in service",
		Buckets: []float64{0.05, 0.1, 0.2, 0.5, 1, 10},
	}, []string{fieldMethod})

	timingBlockMetrics metrics.Histogram = kitprometheus.NewHistogramFrom(prometheus.HistogramOpts{
		Name:    "app_request_block_timing",
		Help:    "timing a request in Blocks",
		Buckets: []float64{0.05, 0.1, 0.2, 0.5, 1, 10},
	}, []string{fieldBlock})

	errorsMetrics metrics.Counter = kitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Name: "request_service_errors",
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
