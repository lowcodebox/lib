package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	timingMetrics metrics.Histogram = kitprometheus.NewHistogramFrom(prometheus.HistogramOpts{
		Name:    "app_request_http_timing",
		Help:    "timing a request in http server",
		Buckets: []float64{0.05, 0.1, 0.2, 0.5, 1, 10},
	}, []string{"method", "http_method"})

	statusCodeMetrics = kitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Name: "app_request_http_status",
		Help: "return status code from http server",
	}, []string{"method", "http_method", "status_code"})
)

// monitoringTiming собирает информацию о времени выполнения запроса
func (*httpserver) monitoringTiming(start time.Time, method, httpMethod string) {
	timingMetrics.
		With("method", method, "http_method", httpMethod).
		Observe(time.Since(start).Seconds())
}

// monitoringStatusCode собирает информацию о статус кодах ответов на запрос
func (*httpserver) monitoringStatusCode(method, httpMethod string, statusCode int) {
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	codeStr := strconv.Itoa(statusCode)
	statusCodeMetrics.
		With("method", method, "http_method", httpMethod, "status_code", codeStr).
		Add(1)
}
