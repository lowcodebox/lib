package lib_test

import (
	"encoding/base64"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Простейшая структура для тестирования DecodeConfig и ConfigLoad (файловая ветка)
type simpleConfig struct {
	A string `toml:"a"`
}

// Для тестирования base64-ветки с массивом таблиц
type tableConfig struct {
	Table []struct {
		A string `toml:"a"`
	} `toml:"Table"`
}

func TestDecodeConfig_Success(t *testing.T) {
	t.Parallel()
	cfg := &simpleConfig{}
	tomlStr := `a = "hello"`
	err := lib.DecodeConfig(tomlStr, cfg)
	assert.NoError(t, err)
	assert.Equal(t, "hello", cfg.A)
}

func TestDecodeConfig_Error(t *testing.T) {
	t.Parallel()
	cfg := &simpleConfig{}
	bad := `a = ` // некорректный TOML
	err := lib.DecodeConfig(bad, cfg)
	assert.Error(t, err)
}

func TestSearchConfigDir_FoundAndNotFound(t *testing.T) {
	t.Parallel()
	// создаём временную структуру каталогов
	root := t.TempDir()
	nested := filepath.Join(root, "dir1", "dir2")
	assert.NoError(t, os.MkdirAll(nested, 0755))

	// создаём файл mycfg.cfg в nested
	wantPath := filepath.Join(nested, "mycfg.cfg")
	assert.NoError(t, os.WriteFile(wantPath, []byte("data"), 0644))

	// должен найти
	found, err := lib.SearchConfigDir(root, "mycfg")
	assert.NoError(t, err)
	assert.Equal(t, wantPath, found)

	// если нет — возвращается пустая строка, без ошибки
	notFound, err := lib.SearchConfigDir(root, "other")
	assert.NoError(t, err)
	assert.Empty(t, notFound)
}

func TestConfigLoad_EmptyConfig(t *testing.T) {
	t.Parallel()
	cfg := &simpleConfig{}
	payload, err := lib.ConfigLoad("", "v1", "h1", cfg)
	assert.Empty(t, payload)
	assert.ErrorIs(t, err, lib.ErrConfig)
}

func TestConfigLoad_FileBranch_Success(t *testing.T) {
	t.Parallel()
	// готовим временный файл с TOML
	tmp := t.TempDir()
	path := filepath.Join(tmp, "cfgfile.cfg")
	tomlStr := `a = "world"`
	assert.NoError(t, os.WriteFile(path, []byte(tomlStr), 0644))

	cfg := &simpleConfig{}
	payload, err := lib.ConfigLoad(path, "ver", "hash", cfg)
	assert.NoError(t, err)
	assert.Equal(t, tomlStr, payload) // payload — это содержимое файла
	assert.Equal(t, "world", cfg.A)   // поле структуры распарсилось
}

func TestConfigLoad_FileBranch_ReadError(t *testing.T) {
	t.Parallel()
	// файл не существует
	cfg := &simpleConfig{}
	_, err := lib.ConfigLoad("no_such_file.cfg", "v", "h", cfg)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unable read configfile"))
}

func TestConfigLoad_Base64Branch_Success(t *testing.T) {
	t.Parallel()
	// подготавливаем TOML-документ в виде массива таблиц, чтобы повторение валидно парсилось
	tomlStr := "[[Table]]\na = \"X\"\n"
	enc := base64.StdEncoding.EncodeToString([]byte(tomlStr))
	// повторим 10 раз, чтобы длина строки >200
	configEnc := strings.Repeat(enc, 10)

	cfg := &tableConfig{}
	payload, err := lib.ConfigLoad(configEnc, "ver2", "hash2", cfg)
	assert.NoError(t, err)

	// payload должен быть исходный TOML, повторённый 10 раз
	expected := strings.Repeat(tomlStr, 10)
	assert.Equal(t, expected, payload)

	// и структура должна содержать 10 элементов
	assert.Len(t, cfg.Table, 10)
}

func TestConfigLoad_Base64Branch_InvalidBase64(t *testing.T) {
	t.Parallel()
	// строка длинная, но невалидная base64
	bad := strings.Repeat("!", 201)
	cfg := &simpleConfig{}
	_, err := lib.ConfigLoad(bad, "v", "h", cfg)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "unable decode to string from base64"))
}
