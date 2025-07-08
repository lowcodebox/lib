//go:build integration
// +build integration

package s3_test

import (
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"strings"
	"testing"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3"
	minio "github.com/minio/minio-go/v7"
)

func TestMinIOIntegration_S3CRUD(t *testing.T) {
	err := localKVStore.Put(ctx, string(s3.ConfigFieldEndpoint), "localhost:9000")
	assert.NoError(t, err)

	minioS3Config, err := s3.NewConfigS3Builder().
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
	assert.NoError(t, err)

	res, err := s3.NewClientMinioBuilder().SetConfig(minioS3Config).Build(context.Background())
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	client := res.Client

	// Create
	bucket := "integ-test-bucket-minio"
	if err := client.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{Region: minioS3Config.Region}); err != nil {
		t.Fatalf("MakeBucket: %v", err)
	}

	// Put
	key, payload := "hello.txt", "Hello, MinIO!"
	if _, err := client.PutObject(
		context.Background(), bucket, key,
		strings.NewReader(payload), int64(len(payload)),
		minio.PutObjectOptions{ContentType: "text/plain"},
	); err != nil {
		t.Fatalf("PutObject: %v", err)
	}

	// Get
	obj, err := client.GetObject(context.Background(), bucket, key, minio.GetObjectOptions{})
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}
	data, _ := io.ReadAll(obj)
	err = obj.Close()
	assert.NoError(t, err)

	if got := string(data); got != payload {
		t.Fatalf("Expected %q, got %q", payload, got)
	}

	// Cleanup
	_ = client.RemoveObject(context.Background(), bucket, key, minio.RemoveObjectOptions{})
	_ = client.RemoveBucket(context.Background(), bucket)
}
