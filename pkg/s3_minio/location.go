package s3_minio

import (
	"context"
	"fmt"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3_wrappers"
	"github.com/minio/minio-go/v7"
	"net/url"
	"strconv"
	"strings"
)

// NewLocation constructs a Location backed by the given MinIO client.
func NewLocation(minioClient *minio.Client) *MinioLocation {
	return &MinioLocation{client: minioClient}
}

// MinioLocation implements Location.
type MinioLocation struct {
	client *minio.Client
}

// Close implements io.Closer. MinIO client has no shutdown step, so no-op.
func (l *MinioLocation) Close() error {
	return nil
}

// CreateContainer creates a new bucket.
func (l *MinioLocation) CreateContainer(name string) (s3_wrappers.Container, error) {
	ctx := context.Background()
	err := l.client.MakeBucket(ctx, name, minio.MakeBucketOptions{})
	if err != nil {
		// if it already exists, return it
		exists, err2 := l.client.BucketExists(ctx, name)
		if err2 == nil && exists {
			return &MinioContainer{client: l.client, bucketName: name}, nil
		}
		return nil, err
	}
	return &MinioContainer{client: l.client, bucketName: name}, nil
}

// Containers lists buckets with prefix, in pages of count, using cursor as an integer offset encoded as string.
func (l *MinioLocation) Containers(prefix, cursor string, count int) ([]s3_wrappers.Container, string, error) {
	ctx := context.Background()
	buckets, err := l.client.ListBuckets(ctx)
	if err != nil {
		return nil, "", err
	}
	// filter by prefix
	var all []s3_wrappers.Container
	for _, b := range buckets {
		if strings.HasPrefix(b.Name, prefix) {
			all = append(all, &MinioContainer{client: l.client, bucketName: b.Name})
		}
	}
	// decode cursor (start index)
	start := 0
	if cursor != "" {
		if i, err := strconv.Atoi(cursor); err == nil {
			start = i
		}
	}
	if start >= len(all) {
		return nil, "", nil
	}
	end := start + count
	if end > len(all) {
		end = len(all)
	}
	page := all[start:end]
	next := ""
	if end < len(all) {
		next = strconv.Itoa(end)
	}
	return page, next, nil
}

// Container gets an existing bucket by name.
func (l *MinioLocation) Container(id string) (s3_wrappers.Container, error) {
	ctx := context.Background()
	exists, err := l.client.BucketExists(ctx, id)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("bucket %q does not exist", id)
	}
	return &MinioContainer{client: l.client, bucketName: id}, nil
}

// RemoveContainer deletes a bucket.
func (l *MinioLocation) RemoveContainer(id string) error {
	ctx := context.Background()
	return l.client.RemoveBucket(ctx, id)
}

// ItemByURL parses a URL of the form http(s)://.../bucketName/objectName and returns an Item.
func (l *MinioLocation) ItemByURL(u *url.URL) (s3_wrappers.Item, error) {
	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid object URL: %s", u.String())
	}
	return &MinioItem{
		client:     l.client,
		bucketName: parts[0],
		objectName: parts[1],
	}, nil
}
