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
