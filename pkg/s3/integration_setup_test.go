//go:build integration
// +build integration

package s3_test

import (
	"os"
	"testing"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
)

var (
	localKVStore = s3.NewLocalKVStore()

	s3Region   = getEnv("AWS_REGION", "us-east-1")
	s3Endpoint = getEnv("S3_ENDPOINT", "http://localhost:9000")
	s3Access   = getEnv("AWS_ACCESS_KEY_ID", "minioadmin")
	s3Secret   = getEnv("AWS_SECRET_ACCESS_KEY", "minioadmin")
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// waitForS3 пробует выполнять ListBuckets до 10 раз с секундным интервалом.
// Если не удалось — паникует.
func waitForS3() {
	s3Config, err := s3.NewConfigS3Builder().
		SetKV(localKVStore).
		SetFieldsToUse([]s3.ConfigField{
			s3.ConfigFieldAuthType,
			s3.ConfigFieldRegion,
			s3.ConfigFieldEndpoint,
			s3.ConfigFieldAccessKeyID,
			s3.ConfigFieldSecretKey,
			s3.ConfigFieldDisableSSL,
		}).
		Build(ctx)
	if err != nil {
		panic("не получается собрать конфиг")
	}
	builder := s3.NewClientAWSBuilder().
		SetConfig(s3Config)

	for i := 0; i < 10; i++ {
		res, err := builder.Build(ctx)
		if err == nil {
			if _, err2 := res.Client.ListBuckets(&awss3.ListBucketsInput{}); err2 == nil {
				return
			}
		}
		time.Sleep(time.Second)
	}
	panic("S3 (MinIO) не стал доступен за 10 секунд")
}

// TestMain один раз подготавливает KV-стор, строит конфиг, ждёт MinIO и запускает все тесты.
func TestMain(m *testing.M) {
	// 1. Инициализируем KV-стор значениями из окружения
	err := s3.InitializeWithMap(ctx, localKVStore, map[s3.ConfigField]string{
		s3.ConfigFieldAuthType:    string(s3.AuthTypeAccessKey),
		s3.ConfigFieldRegion:      s3Region,
		s3.ConfigFieldEndpoint:    s3Endpoint,
		s3.ConfigFieldAccessKeyID: s3Access,
		s3.ConfigFieldSecretKey:   s3Secret,
		s3.ConfigFieldDisableSSL:  "true",
	})
	if err != nil {
		panic("InitializeWithMap: " + err.Error())
	}

	// 2. Строим ConfigS3 из KV-стора

	// 3. Ждём, пока MinIO/S3 станет доступен
	waitForS3()

	// 4. И только потом запускаем тесты
	os.Exit(m.Run())
}
