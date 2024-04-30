package types

import (
	"encoding/json"

	"go.uber.org/zap"
)

func Any(key string, value interface{}) zap.Field {
	j, err := json.Marshal(value)
	if err != nil {
		return zap.Skip()
	}

	return zap.ByteString(key, j)
}
