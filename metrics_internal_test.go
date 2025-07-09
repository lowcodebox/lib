package lib

import (
	"testing"

	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
)

// stub-реализация metrics.Gauge
type gaugeStub struct {
	lastLabels []string
	lastSet    float64
}

func (g *gaugeStub) With(labelValues ...string) metrics.Gauge {
	g.lastLabels = labelValues
	return g
}
func (g *gaugeStub) Add(delta float64) { /*no-op*/ }
func (g *gaugeStub) Set(v float64)     { g.lastSet = v }

func TestSendServiceParamsToMetric(t *testing.T) {
	t.Parallel()
	// Подменяем все package-level gauges на stub-объекты
	stubUID := &gaugeStub{}
	stubName := &gaugeStub{}
	stubVer := &gaugeStub{}
	stubStatus := &gaugeStub{}
	stubPortHTTP := &gaugeStub{}
	stubPID := &gaugeStub{}
	stubReplicas := &gaugeStub{}
	stubPortHTTPS := &gaugeStub{}
	stubDeadTime := &gaugeStub{}
	stubFollower := &gaugeStub{}
	stubPortGRPC := &gaugeStub{}
	stubPortMetrics := &gaugeStub{}

	service_uid = stubUID
	service_name = stubName
	service_version = stubVer
	service_status = stubStatus
	service_port_http = stubPortHTTP
	service_pid = stubPID
	service_replicas = stubReplicas
	service_port_https = stubPortHTTPS
	service_dead_time = stubDeadTime
	service_follower = stubFollower
	service_port_grpc = stubPortGRPC
	service_port_metrics = stubPortMetrics

	// Вызов
	SendServiceParamsToMetric(
		"U", "N", "V",
		"S", "P", "R",
		"PH", "PG", "PM",
		"PHS", "DT", "F",
	)

	// Проверяем, что каждый stub получил свои лейблы и Set(0)
	assert.Equal(t, []string{"value", "U"}, stubUID.lastLabels)
	assert.Equal(t, 0.0, stubUID.lastSet)

	assert.Equal(t, []string{"value", "N"}, stubName.lastLabels)
	assert.Equal(t, 0.0, stubName.lastSet)

	assert.Equal(t, []string{"value", "V"}, stubVer.lastLabels)
	assert.Equal(t, 0.0, stubVer.lastSet)

	assert.Equal(t, []string{"value", "S"}, stubStatus.lastLabels)
	assert.Equal(t, 0.0, stubStatus.lastSet)

	assert.Equal(t, []string{"value", "PH"}, stubPortHTTP.lastLabels)
	assert.Equal(t, 0.0, stubPortHTTP.lastSet)

	assert.Equal(t, []string{"value", "P"}, stubPID.lastLabels)
	assert.Equal(t, 0.0, stubPID.lastSet)

	assert.Equal(t, []string{"value", "R"}, stubReplicas.lastLabels)
	assert.Equal(t, 0.0, stubReplicas.lastSet)

	assert.Equal(t, []string{"value", "PHS"}, stubPortHTTPS.lastLabels)
	assert.Equal(t, 0.0, stubPortHTTPS.lastSet)

	assert.Equal(t, []string{"value", "DT"}, stubDeadTime.lastLabels)
	assert.Equal(t, 0.0, stubDeadTime.lastSet)

	assert.Equal(t, []string{"value", "F"}, stubFollower.lastLabels)
	assert.Equal(t, 0.0, stubFollower.lastSet)

	assert.Equal(t, []string{"value", "PG"}, stubPortGRPC.lastLabels)
	assert.Equal(t, 0.0, stubPortGRPC.lastSet)

	assert.Equal(t, []string{"value", "PM"}, stubPortMetrics.lastLabels)
	assert.Equal(t, 0.0, stubPortMetrics.lastSet)
}
