package types

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func StringMap(key string, val map[string]string) zap.Field {
	res := ""

	for dataKey, dataValue := range val {
		res += fmt.Sprintf("%s:%s;", dataKey, dataValue)
	}

	return zap.Field{Key: key, Type: zapcore.StringType, String: res, Integer: 0, Interface: nil}
}
