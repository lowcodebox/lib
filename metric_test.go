package lib_test

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	lib "git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"github.com/stretchr/testify/assert"
)

func TestNewMetric_ReturnsNonNil(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := lib.NewMetric(ctx, 100*time.Millisecond)
	assert.NotNil(t, m)
}

func TestServiceMetric_QueueAndTPRMetrics(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := lib.NewMetric(ctx, time.Hour) // авто-логгер не мешает

	// имитация +1,+1,+1,-1
	m.SetConnectionIncrement()
	m.SetConnectionIncrement()
	m.SetConnectionIncrement()
	m.SetConnectionDecrement()
	// имитация трёх запросов
	m.SetTimeRequest(10 * time.Millisecond)
	m.SetTimeRequest(20 * time.Millisecond)
	m.SetTimeRequest(30 * time.Millisecond)

	// ждём выполнения горутин и расчёта очереди
	assert.Eventually(t, func() bool {
		m.Generate()
		m.SaveToStash()
		got := m.Get().Queue_AVG
		return math.Abs(float64(got-8.0/3.0)) < 1e-1
	}, 200*time.Millisecond, 10*time.Millisecond,
		"ожидали Queue_AVG≈8/3")

	// после того, как цифра совпала, проверяем остальные поля
	metrics := m.Get()
	assert.Equal(t, float32(0), metrics.Queue_QTL_80)
	assert.Equal(t, 0, metrics.RPS)

	// проверка TPR
	assert.Eventually(t, func() bool {
		m.Generate()
		m.SaveToStash()
		got := m.Get().TPR_AVG_MS
		return math.Abs(float64(got-20000.0)) < 1e-3
	}, 200*time.Millisecond, 10*time.Millisecond,
		"ожидали TPR_AVG_MS≈20000")
}

func TestServiceMetric_MiddlewareIntegration(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := lib.NewMetric(ctx, time.Hour)

	called := false
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(204)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.True(t, called, "underlying handler must be called")
	assert.Equal(t, 204, w.Result().StatusCode)

	// Дождаться работы go-рутинов инкремента и декремента
	time.Sleep(50 * time.Millisecond)

	m.Generate()
	m.SaveToStash()
	metrics := m.Get()

	// после одного входа и выхода queue slice = [1,0], sorted [0,1], sum=1, lQ=2-1=1 → avg=1
	assert.InDelta(t, 1.0, float64(metrics.Queue_AVG), 1e-3)
	assert.Equal(t, 0, metrics.RPS)
}
