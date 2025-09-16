package lib

import (
	"bytes"
	"context"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/internal/utils"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"github.com/stretchr/testify/assert"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	testEndpoint  = utils.GetEnv("MINIO_ENDPOINT", "localhost:9000")
	testAccessKey = utils.GetEnv("MINIO_ACCESS_KEY", "minioadmin")
	testSecretKey = utils.GetEnv("MINIO_SECRET_KEY", "minioadmin")
)

func TestPresignedPostURL(t *testing.T) {
	t.Run("Testing upload for different file types via link", TestPresignedPostURL_DiffFiles)
	t.Run("Testing  file size upload policy", TestPresignedPostURL_Sizes)
	t.Run("Testing security when changing form data", TestPresignedPostURL_FormDataSecurity)
	t.Run("Testing expired policy", TestPresignedPostURL_Duration)
}

func TestPresignedPostURL_DiffFiles(t *testing.T) {
	ctx := context.Background()

	cfg := &models.VFSConfig{
		VfsEndpoint:       testEndpoint,
		VfsAccessKeyID:    testAccessKey,
		VfsSecretKey:      testSecretKey,
		VfsRegion:         "",
		VfsBucket:         "presigned-post-test" + "-" + time.Now().Format("20060102150405"),
		VfsCertCA:         "",
		VfsCDNAccessKeyID: testAccessKey,
		VfsCDNSecretKey:   testSecretKey,
	}

	vfs, err := NewVfs(cfg)
	assert.NoError(t, err)
	defer vfs.Close()

	err = vfs.Connect(ctx)
	assert.NoError(t, err)

	cases := []struct {
		name     string
		filename string
		content  []byte
		policy   UploadPolicy
	}{
		{
			name:     "txt file",
			filename: "file.txt",
			content:  []byte("Просто текст"),
			policy:   UploadPolicy{},
		},
		{
			name:     "jpg image",
			filename: "image.jpg",
			content:  []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01},
			policy:   UploadPolicy{},
		},
		{
			name:     "jpeg with spaces in name",
			filename: "my image.jpeg",
			content:  []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01},
			policy:   UploadPolicy{},
		},
		{
			name:     "png image",
			filename: "image.png",
			content:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			policy:   UploadPolicy{},
		},
		{
			name:     "gif image",
			filename: "animation.gif",
			content:  []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00},
			policy:   UploadPolicy{},
		},
		{
			name:     "webp image",
			filename: "image.webp",
			content:  []byte("RIFF\x00\x00\x00\x00WEBPVP8"),
			policy:   UploadPolicy{},
		},
		{
			name:     "svg vector",
			filename: "vector.svg",
			content:  []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><circle cx="50" cy="50" r="40"/></svg>`),
			policy:   UploadPolicy{},
		},

		{
			name:     "pdf document",
			filename: "document.pdf",
			content:  []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x35, 0x0A, 0x25, 0xD0, 0xD4, 0xC5, 0xD8},
			policy:   UploadPolicy{},
		},
		{
			name:     "csv file",
			filename: "data.csv",
			content:  []byte("id,name,value\n1,test,100\n2,example,200"),
			policy:   UploadPolicy{},
		},
		{
			name:     "json file",
			filename: "config.json",
			content:  []byte(`{"name": "test", "enabled": true, "count": 42}`),
			policy:   UploadPolicy{},
		},
		{
			name:     "xml file",
			filename: "data.xml",
			content:  []byte(`<?xml version="1.0"?><root><item>test</item></root>`),
			policy:   UploadPolicy{},
		},

		{
			name:     "zip archive",
			filename: "archive.zip",
			content:  []byte{0x50, 0x4B, 0x03, 0x04, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00},
			policy:   UploadPolicy{},
		},
		{
			name:     "gzip archive",
			filename: "data.gz",
			content:  []byte{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
			policy:   UploadPolicy{},
		},

		{
			name:     "binary data",
			filename: "data.bin",
			content:  []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
			policy:   UploadPolicy{},
		},
		{
			name:     "file with no extension",
			filename: "justfile",
			content:  []byte("content without extension"),
			policy:   UploadPolicy{},
		},
		{
			name:     "file with multiple dots",
			filename: "archive.tar.gz",
			content:  []byte("compressed data"),
			policy:   UploadPolicy{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := &SignIn{
				Bucket:   cfg.VfsBucket,
				Path:     c.filename,
				Policy:   c.policy,
				Duration: 1 * time.Minute,
			}

			url, formData, err := vfs.GetPresignedPostURL(ctx, in)
			assert.NoError(t, err)

			resp, err := loadFileFromSignedLink(url, formData, c.filename, c.content)
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
				assert.Fail(t, "expected OK got %d", resp.StatusCode)
			}

			file, _, err := vfs.ReadFromBucket(ctx, c.filename, cfg.VfsBucket, false)
			assert.NoError(t, err)
			assert.Equal(t, c.content, file)
		})
	}
}

func TestPresignedPostURL_Sizes(t *testing.T) {
	ctx := context.Background()

	cfg := &models.VFSConfig{
		VfsEndpoint:       testEndpoint,
		VfsAccessKeyID:    testAccessKey,
		VfsSecretKey:      testSecretKey,
		VfsRegion:         "",
		VfsBucket:         "presigned-post-test" + "-" + time.Now().Format("20060102150405"),
		VfsCertCA:         "",
		VfsCDNAccessKeyID: testAccessKey,
		VfsCDNSecretKey:   testSecretKey,
	}

	vfs, err := NewVfs(cfg)
	assert.NoError(t, err)
	defer vfs.Close()

	err = vfs.Connect(ctx)
	assert.NoError(t, err)

	cases := []struct {
		name         string
		filename     string
		content      []byte
		expectStatus int
		policy       UploadPolicy
	}{
		{
			name:         "empty file allowed",
			filename:     "empty.txt",
			content:      []byte(""),
			expectStatus: http.StatusNoContent,
			policy: UploadPolicy{
				MinSize: 0,
				MaxSize: 1024,
			},
		},
		{
			name:         "min size 1KB exact",
			filename:     "exact_1kb.txt",
			content:      []byte(strings.Repeat("a", 1024)),
			expectStatus: http.StatusNoContent,
			policy: UploadPolicy{
				MinSize: 1024,
				MaxSize: 2048,
			},
		},
		{
			name:         "1MB file within limits",
			filename:     "1mb.txt",
			content:      []byte(strings.Repeat("a", 1024*1024)),
			expectStatus: http.StatusNoContent,
			policy: UploadPolicy{
				MinSize: 512 * 1024,
				MaxSize: 1024 * 1024,
			},
		},
		{
			name:         "5MB file within limits",
			filename:     "5mb.txt",
			content:      []byte(strings.Repeat("a", 5*1024*1024)),
			expectStatus: http.StatusNoContent,
			policy: UploadPolicy{
				MinSize: 1,
				MaxSize: 10 * 1024 * 1024,
			},
		},
		{
			name:         "too small should fail",
			filename:     "small.txt",
			content:      []byte(strings.Repeat("a", 500)),
			expectStatus: http.StatusBadRequest,
			policy: UploadPolicy{
				MinSize: 1024,
				MaxSize: 2048,
			},
		},
		{
			name:         "too large should fail",
			filename:     "large.txt",
			content:      []byte(strings.Repeat("a", 3*1024*1024)),
			expectStatus: http.StatusBadRequest,
			policy: UploadPolicy{
				MinSize: 1,
				MaxSize: 2 * 1024 * 1024,
			},
		},
		{
			name:         "no policy small file",
			filename:     "no_policy_small.txt",
			content:      []byte("small"),
			expectStatus: http.StatusNoContent,
			policy:       UploadPolicy{},
		},
		{
			name:         "no policy large file",
			filename:     "no_policy_large.txt",
			content:      []byte(strings.Repeat("a", 3*1024*1024)),
			expectStatus: http.StatusNoContent,
			policy:       UploadPolicy{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := &SignIn{
				Bucket:   cfg.VfsBucket,
				Path:     c.filename,
				Policy:   c.policy,
				Duration: 1 * time.Minute,
			}

			url, formData, gotErr := vfs.GetPresignedPostURL(ctx, in)
			assert.NoError(t, gotErr)

			resp, gotErr := loadFileFromSignedLink(url, formData, c.filename, c.content)
			assert.NoError(t, gotErr)

			assert.Equal(t, c.expectStatus, resp.StatusCode)
		})
	}
}

func TestPresignedPostURL_FormDataSecurity(t *testing.T) {
	ctx := context.Background()

	cfg := &models.VFSConfig{
		VfsEndpoint:       testEndpoint,
		VfsAccessKeyID:    testAccessKey,
		VfsSecretKey:      testSecretKey,
		VfsRegion:         "",
		VfsBucket:         "presigned-post-test" + "-" + time.Now().Format("20060102150405"),
		VfsCertCA:         "",
		VfsCDNAccessKeyID: testAccessKey,
		VfsCDNSecretKey:   testSecretKey,
	}

	vfs, err := NewVfs(cfg)
	assert.NoError(t, err)
	defer vfs.Close()

	err = vfs.Connect(ctx)
	assert.NoError(t, err)

	cases := []struct {
		name             string
		filename         string
		additionFormData map[string]string
		content          []byte
		expectStatus     int
	}{
		{
			name:             "txt file",
			filename:         "file.txt",
			additionFormData: map[string]string{"key": "changed_file.txt"},
			content:          []byte("Просто текст"),
			expectStatus:     http.StatusForbidden,
		},
		{
			name:             "file with ru keys",
			filename:         "Русский файл.txt",
			additionFormData: map[string]string{},
			content:          []byte("Просто текст"),
			expectStatus:     http.StatusNoContent,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := &SignIn{
				Bucket:   cfg.VfsBucket,
				Path:     c.filename,
				Policy:   UploadPolicy{},
				Duration: 1 * time.Minute,
			}

			url, formData, gotErr := vfs.GetPresignedPostURL(ctx, in)
			assert.NoError(t, gotErr)

			for k, v := range c.additionFormData {
				formData[k] = v
			}

			resp, gotErr := loadFileFromSignedLink(url, formData, formData["key"], c.content)
			assert.Equal(t, c.expectStatus, resp.StatusCode)
		})
	}
}

func TestPresignedPostURL_Duration(t *testing.T) {
	ctx := context.Background()

	cfg := &models.VFSConfig{
		VfsEndpoint:       testEndpoint,
		VfsAccessKeyID:    testAccessKey,
		VfsSecretKey:      testSecretKey,
		VfsRegion:         "",
		VfsBucket:         "presigned-post-test" + "-" + time.Now().Format("20060102150405"),
		VfsCertCA:         "",
		VfsCDNAccessKeyID: testAccessKey,
		VfsCDNSecretKey:   testSecretKey,
	}

	vfs, err := NewVfs(cfg)
	assert.NoError(t, err)
	defer vfs.Close()

	err = vfs.Connect(ctx)
	assert.NoError(t, err)

	cases := []struct {
		name         string
		filename     string
		duration     time.Duration
		content      []byte
		expectStatus int
	}{
		{
			name:         "txt file",
			filename:     "file.txt",
			duration:     1 * time.Second,
			content:      []byte("Просто текст"),
			expectStatus: http.StatusForbidden,
		},
		{
			name:         "file with ru keys",
			filename:     "Русский файл.txt",
			duration:     1 * time.Second,
			content:      []byte("Просто текст"),
			expectStatus: http.StatusForbidden,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := &SignIn{
				Bucket:   cfg.VfsBucket,
				Path:     c.filename,
				Policy:   UploadPolicy{},
				Duration: c.duration,
			}

			url, formData, gotErr := vfs.GetPresignedPostURL(ctx, in)
			assert.NoError(t, gotErr)

			time.Sleep(c.duration)

			resp, gotErr := loadFileFromSignedLink(url, formData, c.filename, c.content)
			assert.Equal(t, c.expectStatus, resp.StatusCode)
		})
	}
}

func loadFileFromSignedLink(url string, formData map[string]string, filename string, content []byte) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for k, v := range formData {
		if err := writer.WriteField(k, v); err != nil {
			return nil, err
		}
	}

	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return nil, err
	}

	_, err = part.Write(content)
	if err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return http.DefaultClient.Do(req)
}
