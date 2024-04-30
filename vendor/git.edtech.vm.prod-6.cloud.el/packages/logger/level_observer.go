package logger

import (
	"context"
	"time"

	"go.uber.org/zap/zapcore"
)

type LevelProvider = func(ctx context.Context, curLevel zapcore.Level) zapcore.Level

func SetLevelObserver(ctx context.Context, interval time.Duration, provider LevelProvider) {
	go onceLevelObserver.Do(func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				level.SetLevel(provider(ctx, level.Level()))
			case <-ctx.Done():
				return
			}
		}
	})
}
