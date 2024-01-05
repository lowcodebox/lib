package client

import (
	"context"
	"fmt"
	"time"

	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

func (a *iam) observeLogger(ctx context.Context, start time.Time, method string, err error, arguments ...interface{}) {
	logger.Info(ctx, "timing iam query",
		zap.String("method", method),
		zap.Float64("timing", time.Since(start).Seconds()),
		zap.String("arguments", fmt.Sprint(arguments)),
		zap.Error(err),
	)
}
