package logger

import (
	"context"

	"git.lowcodeplatform.net/packages/logger/types"
	"go.uber.org/zap"
)

const callStackFramesDeep = 3

func prepareFields(fields ...zap.Field) []zap.Field {
	return append([]zap.Field{types.WhereWithDeep(callStackFramesDeep)}, fields...)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	Logger(ctx).Debug(msg, prepareFields(fields...)...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	Logger(ctx).Info(msg, prepareFields(fields...)...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	Logger(ctx).Warn(msg, prepareFields(fields...)...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	Logger(ctx).Error(msg, prepareFields(fields...)...)
}

func Panic(ctx context.Context, msg string, fields ...zap.Field) {
	Logger(ctx).Panic(msg, prepareFields(fields...)...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	Logger(ctx).Fatal(msg, prepareFields(fields...)...)
}
