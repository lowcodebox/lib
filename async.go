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

// Retrier повторяет функцию f в случае наличия ошибки
//
// Параметры:
//
//	retriesMaxCount — максимальное число попыток
//	retriesDelay — минимальная пауза между попытками
//	disableDelayProgression — не увеличивать паузу между попытками
//	f — выполняемая функция
//	isFinalErr [опционально] — проверка на финальную ошибку, после которой нет смысла повторять
func Retrier[T any](
	ctx context.Context,
	retriesMaxCount int,
	retriesDelay time.Duration,
	disableDelayProgression bool,
	f func() (T, error),
	isFinalErr func(e error) bool,
) (res T, err error) {
	for i := 0; i < retriesMaxCount; i++ {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
		}

		if i > 0 {
			if disableDelayProgression {
				time.Sleep(retriesDelay)
			} else {
				time.Sleep(sleepCalc(i, retriesDelay))
			}

			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			default:
			}
		}

		res, err = f()
		if err == nil {
			return
		}
		if isFinalErr != nil && isFinalErr(err) {
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
