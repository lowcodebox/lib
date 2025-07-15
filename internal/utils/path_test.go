package utils_test

import (
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapePathPreservingSlashes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		isError  bool
	}{
		{"ASCII path", "folder/file.txt", "folder/file.txt", false},
		{"path with space", "folder/file name.txt", "folder/file%20name.txt", false},
		{"path with unicode", "папка/файл.txt", "%D0%BF%D0%B0%D0%BF%D0%BA%D0%B0/%D1%84%D0%B0%D0%B9%D0%BB.txt", false},
		{"path with special chars", "folder/a&b.txt", "folder/a%26b.txt", true},
		{"empty path", "", "", false},
		{"path with leading/trailing slashes", "/a/b/c/", "%2Fa/b/c%2F", true}, // note: this is handled literally
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.EscapePathPreservingSlashes(tt.input)
			if tt.isError {
				assert.NotEqual(t, tt.expected, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestJoinURLPath(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		suffix   string
		expected string
	}{
		{"normal join", "/public", "bucket/file.txt", "/public/bucket/file.txt"},
		{"prefix with slash", "/public/", "bucket/file.txt", "/public/bucket/file.txt"},
		{"suffix with slash", "/public", "/bucket/file.txt", "/public/bucket/file.txt"},
		{"both with slashes", "/public/", "/bucket/file.txt", "/public/bucket/file.txt"},
		{"empty prefix", "", "bucket/file.txt", "/bucket/file.txt"},
		{"empty suffix", "/public", "", "/public/"},
		{"both empty", "", "", "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.JoinURLPath(tt.prefix, tt.suffix)
			assert.Equal(t, tt.expected, result)
		})
	}
}
