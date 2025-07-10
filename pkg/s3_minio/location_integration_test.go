//go:build integration
// +build integration

package s3_minio_test

import (
	"bytes"
	"context"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/internal/utils"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3_minio"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"net/url"
	"strconv"
	"testing"
	"time"
)

var (
	testEndpoint  = utils.GetEnv("MINIO_ENDPOINT", "localhost:9000")
	testAccessKey = utils.GetEnv("MINIO_ACCESS_KEY", "minioadmin")
	testSecretKey = utils.GetEnv("MINIO_SECRET_KEY", "minioadmin")
	testUseSSL    = utils.GetEnvBool("MINIO_USE_SSL", false)
)

func setupMinioClient(t *testing.T) *minio.Client {
	client, err := minio.New(testEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(testAccessKey, testSecretKey, ""),
		Secure: testUseSSL,
	})
	assert.NoError(t, err)
	return client
}

func TestMinioLocation_Integration(t *testing.T) {
	ctx := context.Background()

	client := setupMinioClient(t)
	loc := s3_minio.NewLocation(client)

	//ctx := context.Background()
	bucketName := "integration-test-" + time.Now().Format("20060102150405")

	// --- CreateContainer ---
	container, err := loc.CreateContainer(ctx, bucketName)
	assert.NoError(t, err)
	assert.Equal(t, bucketName, container.ID())

	// --- Container ---
	loaded, err := loc.Container(ctx, bucketName)
	assert.NoError(t, err)
	assert.Equal(t, container.ID(), loaded.ID())

	// --- Put item ---
	content := []byte("hello world")
	name := "folder/sample.txt"
	item, err := loaded.Put(ctx, name, bytes.NewReader(content), int64(len(content)), map[string]interface{}{
		"X-Test-Meta": "demo",
	})
	assert.NoError(t, err)
	assert.Equal(t, name, item.ID())

	// --- Get item by ID ---
	fetched, err := loaded.Item(ctx, name)
	assert.NoError(t, err)
	size, _ := fetched.Size(ctx)
	assert.Equal(t, int64(len(content)), size)

	// --- List items with prefix ---
	items, nextCursor, err := loaded.Items(ctx, "folder/", "", 10)
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "", nextCursor)

	// --- URL ---
	u, err := item.URL(ctx)
	assert.NoError(t, err)
	assert.Contains(t, u.String(), "http")

	// --- Metadata ---
	meta, err := item.Metadata(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "demo", meta["X-Test-Meta"])

	// --- ItemByURL ---
	parsedURL, _ := url.Parse(u.String())
	itemByURL, err := loc.ItemByURL(parsedURL)
	assert.NoError(t, err)
	assert.Equal(t, name, itemByURL.Name())

	// --- RemoveItem ---
	err = loaded.RemoveItem(ctx, name)
	assert.NoError(t, err)

	// --- RemoveContainer ---
	err = loc.RemoveContainer(ctx, bucketName)
	assert.NoError(t, err)
}

func TestMinioLocation_ListContainers(t *testing.T) {
	ctx := context.Background()
	client := setupMinioClient(t)
	loc := s3_minio.NewLocation(client)

	prefix := "page-test-" + time.Now().Format("200601021504")

	// --- Создаём 3 test-бакета ---
	var expectedIDs []string
	for i := 1; i <= 3; i++ {
		name := prefix + "-" + strconv.Itoa(i)
		_, err := loc.CreateContainer(ctx, name)
		assert.NoError(t, err)
		expectedIDs = append(expectedIDs, name)
	}

	// --- Пагинация по 2 бакета за раз ---
	cursor := ""
	allBuckets := make([]string, 0)

	for {
		buckets, nextCursor, err := loc.Containers(ctx, prefix, cursor, 2)
		assert.NoError(t, err)

		for _, b := range buckets {
			allBuckets = append(allBuckets, b.ID())
		}

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	// --- Проверка, что все созданные бакеты были найдены ---
	assert.ElementsMatch(t, expectedIDs, allBuckets)

	// --- Чистим за собой ---
	for _, name := range expectedIDs {
		err := loc.RemoveContainer(ctx, name)
		assert.NoError(t, err)
	}
}
