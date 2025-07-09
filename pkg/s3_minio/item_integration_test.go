//go:build integration
// +build integration

package s3_minio_test

import (
	"bytes"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3_minio"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func TestMinioItem_Integration(t *testing.T) {
	client, err := minio.New(testEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(testAccessKey, testSecretKey, ""),
		Secure: testUseSSL,
	})
	assert.NoError(t, err)

	loc := s3_minio.NewLocation(client)
	//ctx := context.Background()
	bucket := "item-test-" + time.Now().Format("20060102150405")

	// Create bucket
	container, err := loc.CreateContainer(bucket)
	assert.NoError(t, err)

	// Upload object
	content := []byte("this is the test content")
	objKey := "test-file.txt"
	_, err = container.Put(objKey, bytes.NewReader(content), int64(len(content)), map[string]interface{}{
		"X-Test-Meta": "test-value",
	})
	assert.NoError(t, err)

	// Get item
	item, err := container.Item(objKey)
	assert.NoError(t, err)

	// --- Size ---
	size, err := item.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)

	// --- Open + Read ---
	r, err := item.Open()
	assert.NoError(t, err)
	defer r.Close()
	readData, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, content, readData)

	// --- ETag ---
	etag, err := item.ETag()
	assert.NoError(t, err)
	assert.NotEmpty(t, etag)

	// --- LastMod ---
	lastMod, err := item.LastMod()
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now(), lastMod, time.Minute)

	// --- Metadata ---
	meta, err := item.Metadata()
	assert.NoError(t, err)
	assert.Equal(t, "test-value", meta["X-Test-Meta"])
	assert.Equal(t, size, meta["Size"])
	assert.Equal(t, etag, meta["ETag"])
	assert.Contains(t, meta["Content-Type"], "application/octet-stream")

	// --- URL ---
	u, err := item.URL()
	assert.NoError(t, err)
	assert.Contains(t, u.String(), objKey)

	// Cleanup
	err = container.RemoveItem(objKey)
	assert.NoError(t, err)
	err = loc.RemoveContainer(bucket)
	assert.NoError(t, err)
}
