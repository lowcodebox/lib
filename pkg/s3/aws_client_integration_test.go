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
	"github.com/aws/aws-sdk-go/aws"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
)

func TestAWSIntegration_S3CRUD(t *testing.T) {

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
	res, err := s3.NewClientAWSBuilder().SetConfig(s3Config).Build(context.Background())
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	client := res.Client

	// Create
	bucket := "integ-test-bucket-aws"
	if _, err := client.CreateBucket(&awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		t.Fatalf("CreateBucket: %v", err)
	}

	// Put
	key, payload := "hello.txt", "Hello, MinIO!"
	if _, err := client.PutObject(&awss3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(payload),
	}); err != nil {
		t.Fatalf("PutObject: %v", err)
	}

	// Get
	out, err := client.GetObject(&awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}
	data, _ := io.ReadAll(out.Body)
	err = out.Body.Close()
	assert.NoError(t, err)

	if got := string(data); got != payload {
		t.Fatalf("Expected %q, got %q", payload, got)
	}

	// Cleanup
	_, _ = client.DeleteObject(&awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	_, _ = client.DeleteBucket(&awss3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
}
