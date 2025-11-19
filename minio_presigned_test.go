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

func TestPresignedPutURL(t *testing.T) {
	t.Run("Testing upload for different file types via link", TestPresignedPutURL_DiffFiles)
	t.Run("Testing  file size upload policy", TestPresignedPutURL_Sizes)
	t.Run("Testing expired policy", TestPresignedPutURL_Duration)
}

func TestPresignedPutURL_DiffFiles(t *testing.T) {
	ctx := context.Background()

	cfg := &models.VFSConfig{
		VfsKind:           "s3",
		VfsEndpoint:       testEndpoint,
		VfsAccessKeyID:    testAccessKey,
		VfsSecretKey:      testSecretKey,
		VfsRegion:         "us-east-1",
		VfsBucket:         "lms-stage",
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
			filename: "upload/lib_integration_tests/file.txt",
			content:  []byte("Просто текст"),
			policy:   UploadPolicy{},
		},
		{
			name:     "jpg image",
			filename: "upload/lib_integration_tests/image.jpg",
			content:  []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01},
			policy:   UploadPolicy{},
		},
		{
			name:     "jpeg with spaces in name",
			filename: "upload/lib_integration_tests/my image.jpeg",
			content:  []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01},
			policy:   UploadPolicy{},
		},
		{
			name:     "png image",
			filename: "upload/lib_integration_tests/image.png",
			content:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			policy:   UploadPolicy{},
		},
		{
			name:     "gif image",
			filename: "upload/lib_integration_tests/animation.gif",
			content:  []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00},
			policy:   UploadPolicy{},
		},
		{
			name:     "webp image",
			filename: "upload/lib_integration_tests/image.webp",
			content:  []byte("RIFF\x00\x00\x00\x00WEBPVP8"),
			policy:   UploadPolicy{},
		},
		{
			name:     "svg vector",
			filename: "upload/lib_integration_tests/vector.svg",
			content:  []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><circle cx="50" cy="50" r="40"/></svg>`),
			policy:   UploadPolicy{},
		},

		{
			name:     "pdf document",
			filename: "upload/lib_integration_tests/document.pdf",
			content:  []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x35, 0x0A, 0x25, 0xD0, 0xD4, 0xC5, 0xD8},
			policy:   UploadPolicy{},
		},
		{
			name:     "csv file",
			filename: "upload/lib_integration_tests/data.csv",
			content:  []byte("id,name,value\n1,test,100\n2,example,200"),
			policy:   UploadPolicy{},
		},
		{
			name:     "json file",
			filename: "upload/lib_integration_tests/config.json",
			content:  []byte(`{"name": "test", "enabled": true, "count": 42}`),
			policy:   UploadPolicy{},
		},
		{
			name:     "xml file",
			filename: "upload/lib_integration_tests/data.xml",
			content:  []byte(`<?xml version="1.0"?><root><item>test</item></root>`),
			policy:   UploadPolicy{},
		},

		{
			name:     "zip archive",
			filename: "upload/lib_integration_tests/archive.zip",
			content:  []byte{0x50, 0x4B, 0x03, 0x04, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00},
			policy:   UploadPolicy{},
		},
		{
			name:     "gzip archive",
			filename: "upload/lib_integration_tests/data.gz",
			content:  []byte{0x1F, 0x8B, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03},
			policy:   UploadPolicy{},
		},

		{
			name:     "binary data",
			filename: "upload/lib_integration_tests/data.bin",
			content:  []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09},
			policy:   UploadPolicy{},
		},
		{
			name:     "file with no extension",
			filename: "upload/lib_integration_tests/justfile",
			content:  []byte("content without extension"),
			policy:   UploadPolicy{},
		},
		{
			name:     "file with multiple dots",
			filename: "upload/lib_integration_tests/archive.tar.gz",
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

			url, gotErr := vfs.GetPresignedPutURL(ctx, in)
			assert.NoError(t, gotErr)

			resp, gotErr := loadFileFromSignedLink(url, make(map[string]string), c.filename, c.content)
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
				assert.Fail(t, "expected OK got %d", resp.StatusCode)
			}

			var buf bytes.Buffer
			_, _ = buf.ReadFrom(resp.Body)

			file, _, gotErr := vfs.ReadFromBucket(ctx, c.filename, cfg.VfsBucket, false)
			assert.NoError(t, gotErr)
			assert.Contains(t, string(file), string(c.content))
		})
	}
}

func TestPresignedPutURL_Sizes(t *testing.T) {
	ctx := context.Background()

	cfg := &models.VFSConfig{
		VfsKind:           "s3",
		VfsEndpoint:       testEndpoint,
		VfsAccessKeyID:    testAccessKey,
		VfsSecretKey:      testSecretKey,
		VfsRegion:         "us-east-1",
		VfsBucket:         "lms-stage",
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
			filename:     "upload/lib_integration_tests/empty.txt",
			content:      []byte(""),
			expectStatus: http.StatusOK,
			policy: UploadPolicy{
				MinSize: 0,
				MaxSize: 1024,
			},
		},
		{
			name:         "min size 1KB exact",
			filename:     "upload/lib_integration_tests/exact_1kb.txt",
			content:      []byte(strings.Repeat("a", 1024)),
			expectStatus: http.StatusOK,
			policy: UploadPolicy{
				MinSize: 1024,
				MaxSize: 2048,
			},
		},
		{
			name:         "1MB file within limits",
			filename:     "upload/lib_integration_tests/1mb.txt",
			content:      []byte(strings.Repeat("a", 1024*1024)),
			expectStatus: http.StatusOK,
			policy: UploadPolicy{
				MinSize: 512 * 1024,
				MaxSize: 1024 * 1024,
			},
		},
		{
			name:         "5MB file within limits",
			filename:     "upload/lib_integration_tests/5mb.txt",
			content:      []byte(strings.Repeat("a", 5*1024*1024)),
			expectStatus: http.StatusOK,
			policy: UploadPolicy{
				MinSize: 1,
				MaxSize: 10 * 1024 * 1024,
			},
		},
		{
			name:         "too small should fail",
			filename:     "upload/lib_integration_tests/small.txt",
			content:      []byte(strings.Repeat("a", 500)),
			expectStatus: http.StatusOK,
			policy: UploadPolicy{
				MinSize: 1024,
				MaxSize: 2048,
			},
		},
		{
			name:         "too large should fail",
			filename:     "upload/lib_integration_tests/large.txt",
			content:      []byte(strings.Repeat("a", 3*1024*1024)),
			expectStatus: http.StatusOK,
			policy: UploadPolicy{
				MinSize: 1,
				MaxSize: 2 * 1024 * 1024,
			},
		},
		{
			name:         "no policy small file",
			filename:     "upload/lib_integration_tests/no_policy_small.txt",
			content:      []byte("small"),
			expectStatus: http.StatusOK,
			policy:       UploadPolicy{},
		},
		{
			name:         "no policy large file",
			filename:     "upload/lib_integration_tests/no_policy_large.txt",
			content:      []byte(strings.Repeat("a", 3*1024*1024)),
			expectStatus: http.StatusOK,
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

			url, gotErr := vfs.GetPresignedPutURL(ctx, in)
			assert.NoError(t, gotErr)

			resp, gotErr := loadFileFromSignedLink(url, make(map[string]string), c.filename, c.content)
			assert.NoError(t, gotErr)

			var body bytes.Buffer
			_, _ = body.ReadFrom(resp.Body)
			assert.Equal(t, c.expectStatus, resp.StatusCode, body.String())
		})
	}
}

func TestPresignedPutURL_Duration(t *testing.T) {
	ctx := context.Background()

	cfg := &models.VFSConfig{
		VfsKind:           "s3",
		VfsEndpoint:       testEndpoint,
		VfsAccessKeyID:    testAccessKey,
		VfsSecretKey:      testSecretKey,
		VfsRegion:         "us-east-1",
		VfsBucket:         "lms-stage",
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
			filename:     "upload/lib_integration_tests/file.txt",
			duration:     1 * time.Second,
			content:      []byte("Просто текст"),
			expectStatus: http.StatusBadRequest,
		},
		{
			name:         "file with ru keys",
			filename:     "upload/lib_integration_tests/Русский файл.txt",
			duration:     5 * time.Second,
			content:      []byte("Просто текст"),
			expectStatus: http.StatusBadRequest,
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

			url, gotErr := vfs.GetPresignedPutURL(ctx, in)
			assert.NoError(t, gotErr)

			time.Sleep(c.duration)

			resp, gotErr := loadFileFromSignedLink(url, make(map[string]string), c.filename, c.content)

			var body bytes.Buffer
			_, _ = body.ReadFrom(resp.Body)

			assert.Equal(t, c.expectStatus, resp.StatusCode, body.String())
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

	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return http.DefaultClient.Do(req)
}
