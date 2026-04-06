package lib

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

func RunAsync(ctx context.Context, fn func()) {
	go func() {
		defer Recover(ctx)
		fn()
	}()
}

// RunAsyncWithLimitExecute выполняет функцию fn в горутине и ожидает её завершения
// с учётом контекста. Возвращает:
//   - nil, если fn успела завершиться до истечения контекста
//   - context.DeadlineExceeded, если контекст истёк
//   - context.Canceled, если контекст был отменён
//   - панику fn, если она произошла (паника преобразуется в ошибку)
func RunAsyncWithLimitExecute(ctx context.Context, fn func()) (err error) {
	// Канал для сигнала завершения fn
	done := make(chan struct{})

	// Канал для перехвата паники
	panicChan := make(chan interface{}, 1)

	// Запускаем fn в отдельной горутине
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Получаем стек для диагностики
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				panicChan <- fmt.Sprintf("panic: %v\nstack: %s", r, buf[:n])
			}
			close(done)
		}()
		fn()
	}()

	// Ожидаем либо завершения fn, либо отмены контекста
	select {
	case <-done:
		// fn успешно завершилась
		return nil

	case panicMsg := <-panicChan:
		// fn упала с паникой
		return fmt.Errorf("%v", panicMsg)

	case <-ctx.Done():
		// Контекст истёк или был отменён
		return ctx.Err()
	}
}

func Recover(ctx context.Context) (flag bool, msg string) {
	recoverErr := recover()
	if recoverErr == nil {
		return false, ""
	}

	stack := debug.Stack()
	pc, file, line, _ := runtime.Caller(2)
	msg = fmt.Sprintf("Recovered panic. file: %s, line: %d, function: %s, error: %s, stack: %s", file, line, runtime.FuncForPC(pc).Name(), recoverErr, stack)

	return true, msg
}

func Retrier[T any](
	retriesMaxCount int,
	retriesDelay time.Duration,
	disableDelayProgression bool,
	f func() (T, error),
) (res T, err error) {
	for i := 0; i < retriesMaxCount; i++ {
		if i > 0 {
			if disableDelayProgression {
				time.Sleep(retriesDelay)
			} else {
				time.Sleep(sleepCalc(i, retriesDelay))
			}
		}

		res, err = f()
		if err == nil {
			return
		}
	}

	return
}

func sleepCalc(ind int, requestRetryMinInterval time.Duration) time.Duration {
	sleep := ind ^ 2
	if sleep > 60 {
		sleep = 60
	}

	return time.Duration(sleep) * requestRetryMinInterval
}
