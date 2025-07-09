package s3_minio

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

// MinioItem implements Item interface backed by MinIO/S3 object.
type MinioItem struct {
	client     *minio.Client
	bucketName string
	objectName string
}

// ID returns the unique object key.
func (i *MinioItem) ID() string {
	return i.objectName
}

// Name returns the object key (you can adapt if needed).
func (i *MinioItem) Name() string {
	return i.objectName
}

// URL returns a presigned GET URL (valid for 1 hour).
func (i *MinioItem) URL() (*url.URL, error) {
	ctx := context.Background()
	// Change expiry if needed
	expiry := time.Hour
	return i.client.PresignedGetObject(ctx, i.bucketName, i.objectName, expiry, nil)
}

// Size returns the object size in bytes.
func (i *MinioItem) Size() (int64, error) {
	ctx := context.Background()
	info, err := i.client.StatObject(ctx, i.bucketName, i.objectName, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

// Open returns a reader to the object.
func (i *MinioItem) Open() (io.ReadCloser, error) {
	ctx := context.Background()
	return i.client.GetObject(ctx, i.bucketName, i.objectName, minio.GetObjectOptions{})
}

// ETag returns the object's ETag (often an MD5 or version marker).
func (i *MinioItem) ETag() (string, error) {
	ctx := context.Background()
	info, err := i.client.StatObject(ctx, i.bucketName, i.objectName, minio.StatObjectOptions{})
	if err != nil {
		return "", err
	}
	return info.ETag, nil
}

// LastMod returns the last modified timestamp.
func (i *MinioItem) LastMod() (time.Time, error) {
	ctx := context.Background()
	info, err := i.client.StatObject(ctx, i.bucketName, i.objectName, minio.StatObjectOptions{})
	if err != nil {
		return time.Time{}, err
	}
	return info.LastModified, nil
}

// Metadata returns user-defined metadata.
func (i *MinioItem) Metadata() (map[string]interface{}, error) {
	ctx := context.Background()
	info, err := i.client.StatObject(ctx, i.bucketName, i.objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	md := make(map[string]interface{})
	for k, v := range info.UserMetadata {
		md[k] = v
	}
	// Optionally include system metadata (Content-Type, etc.)
	md["Content-Type"] = info.ContentType
	md["ETag"] = info.ETag
	md["Size"] = info.Size
	md["LastModified"] = info.LastModified
	return md, nil
}
