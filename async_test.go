package lib_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"github.com/stretchr/testify/assert"
)

func TestRunAsync(t *testing.T) {
	var wg sync.WaitGroup
	ctx := context.Background()

	wg.Add(2)
	// нормальная горутина
	lib.RunAsync(ctx, func() {
		defer wg.Done()
	})
	// горутина, которая падает
	lib.RunAsync(ctx, func() {
		defer wg.Done()
		panic("panic in async")
	})

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// ждем не более секунды, пока обе горутины не отработают
	assert.Eventually(t, func() bool {
		select {
		case <-done:
			return true
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond, "RunAsync tasks did not complete in time")
}

func TestRetrierDelayProgressionCases(t *testing.T) {
	cases := []struct {
		name         string
		retries      int
		baseDelay    time.Duration
		disableProg  bool
		wantMinDelay time.Duration
	}{
		{
			name:         "disableDelayProgression",
			retries:      3,
			baseDelay:    10 * time.Millisecond,
			disableProg:  true,
			wantMinDelay: 0,
		},
		{
			name:        "enableProgression_3retries",
			retries:     3,
			baseDelay:   10 * time.Millisecond,
			disableProg: false,
			// sleepCalc(1,10ms)= (1^2=3)*10ms=30ms
			// sleepCalc(2,10ms)= (2^2=0)*10ms=0
			wantMinDelay: 30 * time.Millisecond,
		},
		{
			name:        "enableProgression_4retries",
			retries:     4,
			baseDelay:   10 * time.Millisecond,
			disableProg: false,
			// sleepCalc(1)=3*10=30, sleepCalc(2)=0, sleepCalc(3)=1*10=10 => total=40ms
			wantMinDelay: 40 * time.Millisecond,
		},
	}

	for _, tc := range cases {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			attempts := 0
			// Функция всегда возвращает ошибку, чтобы Retrier сделал ровно tc.retries попыток
			f := func() (int, error) {
				attempts++
				return 0, errors.New("always fail")
			}

			start := time.Now()
			lib.Retrier(tc.retries, tc.baseDelay, tc.disableProg, f, nil)
			elapsed := time.Since(start)

			assert.Equal(t, tc.retries, attempts,
				"должно быть сделано ровно %d попыток, а сделано %d", tc.retries, attempts)

			// проверяем минимум «спящей» задержки
			assert.GreaterOrEqualf(t, elapsed, tc.wantMinDelay,
				"ожидалось хотя бы %v «спящих» пауз, а прошло %v", tc.wantMinDelay, elapsed)

			// добавляем некоторый запас из 5 базовых задержек, чтобы тест не стал хрупким
			maxExpected := tc.wantMinDelay + 5*tc.baseDelay
			assert.Lessf(t, elapsed, maxExpected,
				"тест слишком затянулся: прошло %v (ожидали < %v)", elapsed, maxExpected)
		})
	}
}
