package lib

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ––– TestSearchConfig –––
// Предполагаем, что SearchConfig(name, startPath)
// ищет файл с именем name, поднимаясь вверх от startPath по дереву директорий.
func TestSearchConfig(t *testing.T) {
	// 1. Создаём временную "корневую" папку
	baseDir := t.TempDir()

	// 2. Внутри делаем nested: base/one/two/three
	deepDir := filepath.Join(baseDir, "one", "two", "three")
	assert.NoError(t, os.MkdirAll(deepDir, 0755))

	// 3. В папке base/one создаём файл конфига myconfig.cfg
	const cfgName = "myconfig"
	cfgFileName := cfgName + ".cfg"
	expectedPath := filepath.Join(baseDir, "one", cfgFileName)
	assert.NoError(t, os.WriteFile(expectedPath, []byte("foo=bar"), 0644))

	// 4. Ищем начиная от baseDir
	found, err := SearchConfig(baseDir, cfgName)
	assert.NoError(t, err)
	assert.Equal(t, expectedPath, found)

	// 5. Если файла нет — должно вернуться "" и err == nil
	notFound, err := SearchConfig(baseDir, "no-such")
	assert.NoError(t, err)
	assert.Empty(t, notFound)
}

// ––– TestTimeParse –––
// Табличный тест для разных форматов строк времени.
func TestTimeParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		nowUTC  bool
		want    time.Time
		wantErr bool
	}{
		{
			name:   "UTC+2",
			input:  "04.04.2024 11:11:11 UTC+2",
			nowUTC: true,
			// 11:11@UTC+2 == 09:11 UTC
			want:    time.Date(2024, 4, 4, 9, 11, 11, 0, time.UTC),
			wantErr: false,
		},
		{
			name:   "MSK minus 1d3h",
			input:  "04.04.2024 11:11:11 MSK - 1d3h",
			nowUTC: true,
			// 11:11@MSK(UTC+3) == 08:11 UTC, минус 1d3h => 05:11 UTC предыдущего дня
			want:    time.Date(2024, 4, 3, 5, 11, 11, 0, time.UTC),
			wantErr: false,
		},
		{
			name:   "MSK minus 1d minus 3h (с пробелами)",
			input:  "04.04.2024 11:11:11 MSK - 1d - 3h",
			nowUTC: true,
			// То же самое, ожидаем 05:11 UTC предыдущего дня
			want:    time.Date(2024, 4, 3, 5, 11, 11, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "bad format",
			input:   "not a date",
			nowUTC:  true,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := TimeParse(tc.input, tc.nowUTC)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.True(
				t,
				got.Equal(tc.want),
				"expected %v, got %v",
				tc.want, got,
			)
		})
	}
}
