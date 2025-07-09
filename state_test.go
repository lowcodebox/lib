package lib_test

import (
	"math"
	"runtime"
	"testing"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"github.com/stretchr/testify/assert"
)

func TestStateHostTick_EmptyHost(t *testing.T) {
	t.Parallel()
	sh := &lib.StateHost{}

	// Перед Tick всё должно быть нулём
	assert.Equal(t, 0.0, sh.PercentageMemory)
	assert.Equal(t, 0, sh.Goroutines)

	sh.Tick()

	// После Tick PercentageMemory установится в [0..100]
	assert.GreaterOrEqual(t, sh.PercentageMemory, 0.0, "PercentageMemory ≥ 0")
	assert.LessOrEqual(t, sh.PercentageMemory, 100.0, "PercentageMemory ≤ 100")
	// Поскольку используется math.Round, значение должно быть целым
	assert.Equal(t, math.Round(sh.PercentageMemory), sh.PercentageMemory,
		"PercentageMemory — округлённое целое")

	// Goroutines должно совпадать с runtime.NumGoroutine()
	assert.Equal(t, runtime.NumGoroutine(), sh.Goroutines,
		"Goroutines должно = runtime.NumGoroutine()")
}

func TestStateHostTick_PreservesOtherFields(t *testing.T) {
	// Заполняем все поля нетривиальными значениями
	sh := &lib.StateHost{
		PercentageCPU:  12.34,
		PercentageDisk: 56.78,
		TotalCPU:       1,
		TotalMemory:    2,
		TotalDisk:      3,
		UsedCPU:        4,
		UsedMemory:     5,
		UsedDisk:       6,
		Goroutines:     999,
	}

	sh.Tick()

	// Убедимся, что Tick затронул только PercentageMemory и Goroutines,
	// а остальные поля остались без изменений.
	assert.Equal(t, 12.34, sh.PercentageCPU, "PercentageCPU не должен меняться")
	assert.Equal(t, 56.78, sh.PercentageDisk, "PercentageDisk не должен меняться")
	assert.Equal(t, 1.0, sh.TotalCPU)
	assert.Equal(t, 2.0, sh.TotalMemory)
	assert.Equal(t, 3.0, sh.TotalDisk)
	assert.Equal(t, 4.0, sh.UsedCPU)
	assert.Equal(t, 5.0, sh.UsedMemory)
	assert.Equal(t, 6.0, sh.UsedDisk)

	// Проверяем новые значения
	assert.GreaterOrEqual(t, sh.PercentageMemory, 0.0,
		"PercentageMemory ≥ 0")
	assert.LessOrEqual(t, sh.PercentageMemory, 100.0,
		"PercentageMemory ≤ 100")
	assert.Equal(t, runtime.NumGoroutine(), sh.Goroutines,
		"Goroutines должно = runtime.NumGoroutine()")
}
