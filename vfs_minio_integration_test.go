//go:build integration
// +build integration

package lib_test

import (
	"context"
	"fmt"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/internal/utils"
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

var (
	testEndpoint  = utils.GetEnv("MINIO_ENDPOINT", "localhost:9000")
	testAccessKey = utils.GetEnv("MINIO_ACCESS_KEY", "minioadmin")
	testSecretKey = utils.GetEnv("MINIO_SECRET_KEY", "minioadmin")
	testUseSSL    = utils.GetEnvBool("MINIO_USE_SSL", false)

	testSecureEndpoint  = utils.GetEnv("MINIO_SECURE_ENDPOINT", "localhost:9443")
	testSecureAccessKey = utils.GetEnv("MINIO_SECURE_ACCESS_KEY", "minioadmin")
	testSecureSecretKey = utils.GetEnv("MINIO_SECURE_SECRET_KEY", "minioadmin")
	testSecureUseSSL    = utils.GetEnvBool("MINIO_SECURE_USE_SSL", true)
)

func TestVfsMinio_WriteReadDelete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		endpoint  string
		accessKey string
		secretKey string
		useSSL    bool
		caCert    string
	}{
		{
			name:      "insecure",
			endpoint:  testEndpoint,
			accessKey: testAccessKey,
			secretKey: testSecretKey,
			useSSL:    testUseSSL,
			caCert:    "",
		},
		{
			name:      "secure",
			endpoint:  testSecureEndpoint,
			accessKey: testSecureAccessKey,
			secretKey: testSecureSecretKey,
			useSSL:    testSecureUseSSL,
			caCert:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &lib.VfsConfig{
				Endpoint:    tt.endpoint,
				AccessKeyID: tt.accessKey,
				SecretKey:   tt.secretKey,
				Region:      "",
				Bucket:      "vfs-test-" + tt.name + "-" + time.Now().Format("20060102150405"),
				UseSSL:      tt.useSSL,
				CACert:      tt.caCert,
			}

			vfs, err := lib.NewVfs(cfg)
			assert.NoError(t, err)
			defer vfs.Close()

			err = vfs.Connect(ctx)
			assert.NoError(t, err)

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "example.txt")
			originalContent := []byte("hello from integration test!")
			err = os.WriteFile(tmpFile, originalContent, 0644)
			assert.NoError(t, err)

			data, err := os.ReadFile(tmpFile)
			assert.NoError(t, err)

			minioPath := "test-folder/example.txt"
			err = vfs.Write(ctx, minioPath, data)
			assert.NoError(t, err)

			readData, mimeType, err := vfs.Read(ctx, minioPath, false)
			assert.NoError(t, err)
			assert.Equal(t, data, readData)
			assert.Contains(t, mimeType, "text/")

			err = vfs.Delete(ctx, minioPath)
			assert.NoError(t, err)
		})
	}
}

func TestVfsMinio_ItemAndList(t *testing.T) {

	ctx := context.Background()

	tests := []struct {
		name      string
		endpoint  string
		accessKey string
		secretKey string
		useSSL    bool
		caCert    string
	}{
		{
			name:      "insecure",
			endpoint:  testEndpoint,
			accessKey: testAccessKey,
			secretKey: testSecretKey,
			useSSL:    testUseSSL,
			caCert:    "",
		},
		{
			name:      "secure",
			endpoint:  testSecureEndpoint,
			accessKey: testSecureAccessKey,
			secretKey: testSecureSecretKey,
			useSSL:    testSecureUseSSL,
			caCert:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &lib.VfsConfig{
				Endpoint:    tt.endpoint,
				AccessKeyID: tt.accessKey,
				SecretKey:   tt.secretKey,
				Region:      "",
				Bucket:      "item-list-test-" + tt.name + "-" + time.Now().Format("20060102150405"),
				UseSSL:      tt.useSSL,
				CACert:      tt.caCert,
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
		})
	}
}

func TestVfsMinio_Proxy(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		endpoint  string
		accessKey string
		secretKey string
		useSSL    bool
		caCert    string
	}{
		{
			name:      "insecure",
			endpoint:  testEndpoint,
			accessKey: testAccessKey,
			secretKey: testSecretKey,
			useSSL:    testUseSSL,
			caCert:    "",
		},
		{
			name:      "secure",
			endpoint:  testSecureEndpoint,
			accessKey: testSecureAccessKey,
			secretKey: testSecureSecretKey,
			useSSL:    testSecureUseSSL,
			caCert:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originCfg := &lib.VfsConfig{
				Endpoint:    tt.endpoint,
				AccessKeyID: tt.accessKey,
				SecretKey:   tt.secretKey,
				Region:      "",
				Bucket:      "proxy-test-" + tt.name + time.Now().Format("20060102150405"),
				UseSSL:      tt.useSSL,
				CACert:      tt.caCert,
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
		})
	}
}

const (
	caTestingCert = `
-----BEGIN CERTIFICATE-----
MIIFhzCCA2+gAwIBAgIUaV/Pnk+OiHgmyfDHDEciP77jCiwwDQYJKoZIhvcNAQEL
BQAwUjELMAkGA1UEBhMCUlUxDTALBgNVBAgMBFRlc3QxDTALBgNVBAcMBFRlc3Qx
EDAOBgNVBAoMB1Rlc3RPcmcxEzARBgNVBAMMClRlc3RSb290Q0EwIBcNMjUwNzEw
MTExMzI1WhgPMjEyNTA2MTYxMTEzMjVaMFIxCzAJBgNVBAYTAlJVMQ0wCwYDVQQI
DARUZXN0MQ0wCwYDVQQHDARUZXN0MRAwDgYDVQQKDAdUZXN0T3JnMRMwEQYDVQQD
DApUZXN0Um9vdENBMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAqDYB
WmVzt0/kBuJTwAEWB7oaXdxx3+1P6IUt8FWDeVJF/zc8aMFQqeprSG48Lg0TWlKB
tzC7ProFkc6wsP+BiHD+sSr1JIMPLK475IFCruEiSUQZ8V5Dy+dqGiuRuczJrnJt
CUsRAU4WTQt7IA7p/Qx3IFmj2puKb2bGV2W5U36qsCNHPOoQ2hyAb8+T+gxDEbfe
OZ3afC4ZFF87xoUEid/Hz29/VUFPQn8IoHmvuKh9Js6XV5sA6R/gs4IHvmr/7pDk
lDHWQroH6V19xaeicTzjf8DMSAXmN6rMq5fItEO0pak7oOAQz+AH928DJMWT9Lxp
dbMo5G1AY8Y96xfii8iNMj3aF23LKSQBATQmNRP/tmz/Oa1VC9A5bVvvSKXtvisY
p0S2nDpt1JVsQgtUXHm6ob4jUN+sYUtlFOCwPgn8NqUT2m5e9sTWhX6LZ+Ri29Us
EVWG7rmEbUJhxCUcjiD+6sGt9VWgNsyOwE7WW1sOd6YMDMV2pnridZKGEwgqRY+N
sQXCX1UbMfPzQW2EgJSuYgMti0AN+eUG0KUK33W9mACVQhX9YvMBgv+glFAP9lKe
XPGhaferVJSh9+hohPplYHyVBVxYNhdXoqFxNUUyO3b1jd7FydJxPrmXKjQgmFNA
qdRJqvJ2QSv9Sjo4iXjhV+aplSiwIyGesX+NjfUCAwEAAaNTMFEwHQYDVR0OBBYE
FGkQW0LdEK5i6pNSFmn4grBWXkv5MB8GA1UdIwQYMBaAFGkQW0LdEK5i6pNSFmn4
grBWXkv5MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggIBAD33+WV9
NZwRZ5b5MqYSqVBe0YrcyO3SpOgXnR92h0Pao71ffYOhQG3QI34rgro2PwSqhUum
zY1pnWMk6cy9ojLzuDnvHkqTk04dF62GORPQjBW2NX8QpEVaFPfvuye8rbl04tMo
HOx+YWNX378Xe7ww+zVskJ3/88B5NXWvfE2FSi5YzQWfXnB03Ds02Bp2530Fr7eC
HDZG4IO8UccfLnul2QBTYSOQ9Md/NKuMFQcbaCVzPofLXgaoxZFfc7XxJ43ru6fz
Otp1S7QjjoiDHL9j7tt1OPdVf3Pev1SGaWijvef0GoE8XbfO/XgZNf1rid+vChix
XphaitDwsXwxpBBWA/Ur2SajjJlXKwl5xBVh4LWoirXhdNHYZQnqNgdzdW3c92tF
LCZl8NbszlwqwRYb2/vP6xZYesXOXIlenCXFNL+D17iiFkGgwrMx4yuhD0hp/Pbe
YS6i7maOKpftYlgtY59gZGchpPRjzuEhWq3oqF++5j9QXS9p0VMJOFEftYhw7fDE
mtkzyAxsbh00YrsyqjXsbEZYRCy5Ign4GHudMOlTne3m7f9AIfVchyM36Jd6dxQ7
faHfG83A/95LWSJZ65YD4zfjH8r2LkSxBB5m6sHZ8a9I+FMD5GZ2GLVlp3QQE/CP
TVbl2CPibe2Vthh0PZhpqlqKbovKTvNNX5iX
-----END CERTIFICATE-----
`
)
