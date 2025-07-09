package lib_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"github.com/stretchr/testify/assert"
)

func TestVfsMinio_WriteReadDelete(t *testing.T) {
	ctx := context.Background()

	cfg := &lib.VfsConfig{
		Endpoint:    "localhost:9000",
		AccessKeyID: "minioadmin",
		SecretKey:   "minioadmin",
		Region:      "",
		Bucket:      "vfs-test-" + time.Now().Format("20060102150405"),
		UseSSL:      false,
	}

	// Создаём VFS
	vfs, err := lib.NewVfs(cfg)
	assert.NoError(t, err)
	defer vfs.Close()

	// Connect & Ensure bucket exists
	err = vfs.Connect()
	assert.NoError(t, err)

	// Создаём временный файл
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "example.txt")
	originalContent := []byte("hello from integration test!")
	err = os.WriteFile(tmpFile, originalContent, 0644)
	assert.NoError(t, err)

	// Читаем его как []byte
	data, err := os.ReadFile(tmpFile)
	assert.NoError(t, err)

	// Записываем файл в MinIO
	minioPath := "test-folder/example.txt"
	err = vfs.Write(ctx, minioPath, data)
	assert.NoError(t, err)

	//// Читаем его обратно
	readData, mimeType, err := vfs.Read(ctx, minioPath, false)
	assert.NoError(t, err)
	assert.Equal(t, data, readData)
	assert.Contains(t, mimeType, "text/") // MIME может быть text/plain

	// Удаляем файл
	err = vfs.Delete(ctx, minioPath)
	assert.NoError(t, err)
}

func TestVfsMinio_Proxy(t *testing.T) {
	t.Skip()
	ctx := context.Background()

	cfg := &lib.VfsConfig{
		Endpoint:    "localhost:9000",
		AccessKeyID: "minioadmin",
		SecretKey:   "minioadmin",
		Region:      "",
		Bucket:      "proxy-test-" + time.Now().Format("20060102150405"),
		UseSSL:      false,
	}

	// создаём VFS
	vfs, err := lib.NewVfs(cfg)
	assert.NoError(t, err)
	defer vfs.Close()

	err = vfs.Connect()
	assert.NoError(t, err)

	// записываем файл
	objectPath := "files/sample.txt"
	content := []byte("proxied content")
	err = vfs.Write(ctx, objectPath, content)
	assert.NoError(t, err)

	// создаём прокси
	proxyHandler, err := vfs.Proxy("/public", "")
	assert.NoError(t, err)

	server := httptest.NewServer(proxyHandler)
	defer server.Close()

	// делаем запрос через прокси
	resp, err := http.Get(server.URL + "/public/" + cfg.Bucket + "/" + objectPath)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, content, body)
}
