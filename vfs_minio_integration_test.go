//go:build integration
// +build integration

package lib_test

import (
	"context"
	"fmt"
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
	err = vfs.Connect(ctx)
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

func TestVfsMinio_ItemAndList(t *testing.T) {
	ctx := context.Background()

	cfg := &lib.VfsConfig{
		Endpoint:    "localhost:9000",
		AccessKeyID: "minioadmin",
		SecretKey:   "minioadmin",
		Region:      "",
		Bucket:      "item-list-test-" + time.Now().Format("20060102150405"),
		UseSSL:      false,
	}

	vfs, err := lib.NewVfs(cfg)
	assert.NoError(t, err)
	defer vfs.Close()

	err = vfs.Connect(ctx)
	assert.NoError(t, err)

	// Записываем несколько файлов
	filesToWrite := map[string][]byte{
		"folder1/file1.txt": []byte("content1"),
		"folder1/file2.txt": []byte("content2"),
		"folder1/file3.txt": []byte("content3"),
	}

	for path, content := range filesToWrite {
		err := vfs.Write(ctx, path, content)
		assert.NoError(t, err)
	}

	// Проверяем метод Item
	item, err := vfs.Item(ctx, "folder1/file1.txt")
	assert.NoError(t, err)
	assert.Equal(t, "folder1/file1.txt", item.ID())
	itemSize, err := item.Size(ctx)
	assert.NoError(t, err)
	assert.EqualValues(t, len(filesToWrite["folder1/file1.txt"]), itemSize)

	// Проверяем метод List
	listedItems, err := vfs.List(ctx, "folder1/", 100)
	assert.NoError(t, err)
	assert.Len(t, listedItems, len(filesToWrite))

	listedKeys := make(map[string]bool)
	for _, item := range listedItems {
		listedKeys[item.ID()] = true
	}

	for expectedKey := range filesToWrite {
		assert.True(t, listedKeys[expectedKey], "missing expected key in list: %s", expectedKey)
	}

	// Очистка
	for path := range filesToWrite {
		err := vfs.Delete(ctx, path)
		assert.NoError(t, err)
	}
}

func TestVfsMinio_Proxy(t *testing.T) {
	ctx := context.Background()

	originCfg := &lib.VfsConfig{
		Endpoint:    "localhost:9000",
		AccessKeyID: "minioadmin",
		SecretKey:   "minioadmin",
		Region:      "",
		Bucket:      "proxy-test-" + time.Now().Format("20060102150405"),
		UseSSL:      false,
	}

	// создаём основной VFS
	vfsOrigin, err := lib.NewVfs(originCfg)
	assert.NoError(t, err)
	defer vfsOrigin.Close()

	err = vfsOrigin.Connect(ctx)
	assert.NoError(t, err)

	// записываем файл напрямую
	objectPath := "files/sample.txt"
	expectedContent := []byte("proxied content")
	err = vfsOrigin.Write(ctx, objectPath, expectedContent)
	assert.NoError(t, err)

	// создаём прокси-обработчик
	proxyHandler, err := vfsOrigin.Proxy("/public/", "/")
	assert.NoError(t, err)

	// поднимаем HTTP-сервер
	testServer := httptest.NewServer(http.StripPrefix("/public", proxyHandler))
	defer testServer.Close()

	// формируем URL, по которому будет доступен файл через прокси
	proxyURL := fmt.Sprintf("%s/public/%s/%s", testServer.URL, originCfg.Bucket, objectPath)

	// отправляем GET-запрос через прокси
	resp, err := http.Get(proxyURL)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	// сравниваем с ожидаемым содержимым
	assert.Equal(t, expectedContent, body)
}
