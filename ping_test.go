package lib

import (
	"os"
	"runtime"
	"testing"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"github.com/stretchr/testify/assert"
)

// helper для сброса глобального состояния перед каждым тестом
func resetGlobals() {
	pongObj = models.PongObj{}
	pingConf = models.PingConfig{}
	pingConfOld = models.PingConfigOld{}
	configName = ""
	// startTime оставим текущее, тесты при необходимости сами переустанавливают его
}

func TestPingInitialFields(t *testing.T) {
	t.Parallel()
	resetGlobals()

	// установим фиксированный старт для детерминизма в StartTime
	t0 := time.Now().Add(-5 * time.Second)
	startTime = t0

	got := Ping()
	// ReplicaID всегда должно быть сгенерировано
	assert.NotEmpty(t, got.ReplicaID)
	// Status и Pid
	assert.Equal(t, "run", got.Status)
	assert.Equal(t, os.Getpid(), got.Pid)
	// StartTime берётся из нашей переменной
	assert.Equal(t, t0, got.StartTime)
	// Uptime — это time.Since(startTime).String(), всегда не пустая строка и parseable
	assert.NotEmpty(t, got.Uptime)
	_, err := time.ParseDuration(got.Uptime)
	assert.NoError(t, err)
	// ОС и архитектура совпадают с тем, на чем запущен тест
	assert.Equal(t, runtime.GOOS, got.OS)
	assert.Equal(t, runtime.GOARCH, got.Arch)
}

func TestPingSubsequentUptimeIncreases(t *testing.T) {
	t.Parallel()
	resetGlobals()
	// для удобства короткий старт
	startTime = time.Now().Add(-100 * time.Millisecond)

	first := Ping()
	d1, err := time.ParseDuration(first.Uptime)
	assert.NoError(t, err)

	// немного подождём, чтобы Uptime точно увеличился
	time.Sleep(20 * time.Millisecond)

	second := Ping()
	d2, err := time.ParseDuration(second.Uptime)
	assert.NoError(t, err)

	assert.GreaterOrEqual(t, d2, d1, "ожидали, что Uptime на втором Ping не меньше первого")
}

func TestPingUsesConfigNameForUid(t *testing.T) {
	t.Parallel()
	resetGlobals()
	configName = "mycfg"
	got := Ping()
	assert.Equal(t, "mycfg", got.Uid, "если pingConf.Uid и pingConfOld.DataUid пусты, используется configName")
}

func TestPingSplitsOldDomainToProjectAndName(t *testing.T) {
	t.Parallel()
	resetGlobals()
	// имитируем устаревший формат domain
	pingConfOld.Domain = "projX/nameY"
	// configName не должен влиять
	configName = "zzz"
	got := Ping()
	assert.Equal(t, "projX", got.Project, "должно взять часть до '/' из pingConfOld.Domain")
	assert.Equal(t, "nameY", got.Name, "должно взять часть после '/' из pingConfOld.Domain")
}

func TestSetPongFieldsModifiesAndCallsPingOnce(t *testing.T) {
	t.Parallel()
	resetGlobals()
	// нам нужен фиксированный старт, чтобы SetPongFields внутри Ping выставил StartTime
	t0 := time.Now().Add(-1 * time.Minute)
	startTime = t0

	// перед вызовом pongObj ещё не инициализирован
	assert.Empty(t, pongObj.ReplicaID)

	// вызываем SetPongFields — внутри него Ping(), а затем наш callback
	SetPongFields(func(p *models.PongObj) {
		p.Status = "patched"
		p.Project = "P123"
	})

	// после этого pongObj уже инициализирован и в него записались наши патчи
	assert.Equal(t, "patched", pongObj.Status)
	assert.Equal(t, "P123", pongObj.Project)
	// и StartTime сохранился из t0
	assert.Equal(t, t0, pongObj.StartTime)
}

func TestSetPongFieldsNilDoesNothing(t *testing.T) {
	t.Parallel()
	resetGlobals()
	// pongObj ещё пуст
	assert.Empty(t, pongObj.ReplicaID)
	// вызов с nil не должен вызвать ни Panic, ни Ping
	SetPongFields(nil)
	assert.Empty(t, pongObj.ReplicaID, "при f==nil pongObj не изменяется и Ping не вызывается")
}
