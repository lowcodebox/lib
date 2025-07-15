package utils_test

import (
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/internal/utils"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	// Set test env var
	_ = os.Setenv("TEST_STRING_KEY", "hello")

	t.Run("existing key returns value", func(t *testing.T) {
		val := utils.GetEnv("TEST_STRING_KEY", "default")
		assert.Equal(t, "hello", val)
	})

	t.Run("missing key returns default", func(t *testing.T) {
		val := utils.GetEnv("NON_EXISTENT_KEY", "default")
		assert.Equal(t, "default", val)
	})
}

func TestGetEnvBool(t *testing.T) {
	_ = os.Setenv("BOOL_TRUE_1", "true")
	_ = os.Setenv("BOOL_TRUE_2", "1")
	_ = os.Setenv("BOOL_FALSE", "false")
	_ = os.Setenv("BOOL_GARBAGE", "xyz")

	tests := []struct {
		name       string
		key        string
		defaultVal bool
		expected   bool
	}{
		{"true string", "BOOL_TRUE_1", false, true},
		{"true 1", "BOOL_TRUE_2", false, true},
		{"false string", "BOOL_FALSE", true, false},
		{"invalid value falls back to false", "BOOL_GARBAGE", false, false},
		{"missing key returns default (true)", "MISSING_BOOL", true, true},
		{"missing key returns default (false)", "MISSING_BOOL", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := utils.GetEnvBool(tt.key, tt.defaultVal)
			assert.Equal(t, tt.expected, val)
		})
	}
}
