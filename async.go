package lib

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
)

func RunAsync(ctx context.Context, fn func()) {
	go func() {
		defer Recover(ctx)
		fn()
	}()
}

func Recover(ctx context.Context) bool {
	recoverErr := recover()
	if recoverErr == nil {
		return false
	}

	stack := debug.Stack()
	pc, file, line, _ := runtime.Caller(2)
	msg := fmt.Sprintf("Recovered panic. file: %s, line: %d, function: %s, error: %s, stack: %s", file, line, runtime.FuncForPC(pc).Name(), recoverErr, stack)

	// TODO заменить на логгер
	fmt.Println(msg)

	return true
}
