package httpserver

import (
	"strconv"
	"time"

	"github.com/go-kit/kit/metrics"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	timingMetrics metrics.Histogram = kitprometheus.NewSummaryFrom(prometheus.SummaryOpts{
		Name:       "app_request_http_timing",
		Help:       "timing a request in http server",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001},
	}, []string{"method", "http_method", "request_uri"})

	codeMetrics = kitprometheus.NewCounterFrom(prometheus.CounterOpts{
		Name: "app_request_http_status",
		Help: "return status code from http server",
	}, []string{"method", "http_method", "request_uri", "status_code"})
)

// observeMetric собирает информацию о времени выполнения запроса
func (*httpserver) observeMetric(start time.Time, method, httpMethod, requestURI string) {
	timingMetrics.
		With("method", method, "http_method", httpMethod, "request_uri", requestURI).
		Observe(time.Since(start).Seconds())
}

// statusCodeMetric собирает информацию о статус кодах ответов на запрос
func (*httpserver) statusCodeMetric(method, httpMethod, requestURI string, statusCode int) {
	codeStr := strconv.Itoa(statusCode)
	codeMetrics.
		With("method", method, "http_method", httpMethod, "request_uri", requestURI, "status_code", codeStr).
		Add(1)
}
