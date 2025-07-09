package s3_minio

import (
	"context"
	"fmt"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3_wrappers"
	"io"
	"strconv"

	"github.com/minio/minio-go/v7"
)

type MinioContainer struct {
	client     *minio.Client
	bucketName string
}

func (c *MinioContainer) ID() string {
	return c.bucketName
}

func (c *MinioContainer) Name() string {
	return c.bucketName
}

func (c *MinioContainer) Item(id string) (s3_wrappers.Item, error) {
	ctx := context.Background()
	_, err := c.client.StatObject(ctx, c.bucketName, id, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &MinioItem{
		client:     c.client,
		bucketName: c.bucketName,
		objectName: id,
	}, nil
}

// Items returns a paginated list of items using offset-based paging (not S3 continuation token).
func (c *MinioContainer) Items(prefix, cursor string, count int) ([]s3_wrappers.Item, string, error) {
	ctx := context.Background()
	start := 0
	if cursor != "" {
		var err error
		start, err = strconv.Atoi(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("invalid cursor: %w", err)
		}
	}

	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	itemCh := c.client.ListObjects(ctx, c.bucketName, opts)

	var (
		items     []s3_wrappers.Item
		collected int
		skipped   int
	)

	for obj := range itemCh {
		if obj.Err != nil {
			return nil, "", obj.Err
		}

		if skipped < start {
			skipped++
			continue
		}

		items = append(items, &MinioItem{
			client:     c.client,
			bucketName: c.bucketName,
			objectName: obj.Key,
		})

		collected++
		if collected >= count {
			break
		}
	}

	var nextCursor string
	if collected == count {
		nextCursor = strconv.Itoa(start + collected)
	}

	return items, nextCursor, nil
}

func (c *MinioContainer) RemoveItem(id string) error {
	ctx := context.Background()
	return c.client.RemoveObject(ctx, c.bucketName, id, minio.RemoveObjectOptions{})
}

func (c *MinioContainer) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (s3_wrappers.Item, error) {
	ctx := context.Background()

	userMeta := make(map[string]string, len(metadata))
	for k, v := range metadata {
		userMeta[k] = fmt.Sprint(v)
	}

	opts := minio.PutObjectOptions{
		UserMetadata: userMeta,
	}
	_, err := c.client.PutObject(ctx, c.bucketName, name, r, size, opts)
	if err != nil {
		return nil, err
	}

	return &MinioItem{
		client:     c.client,
		bucketName: c.bucketName,
		objectName: name,
	}, nil
}
