// lib/vfs_minio.go
package lib

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3_minio"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3_wrappers"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const (
	minioVfsKind = "s3"
)

type vfsMinio struct {
	client    *minio.Client
	location  s3_wrappers.Location
	container s3_wrappers.Container
	config    *VfsConfig
}

// VfsConfig — конфиг для MinIO/S3
type VfsConfig struct {
	Endpoint    string // host:port или полный URL
	AccessKeyID string
	SecretKey   string
	Region      string
	Bucket      string
	UseSSL      bool   // https или http
	Comma       string // заменитель для "/"
	CACert      string // если нужен кастомный CA — можно передать, но в этом примере не используется
}

// Validate проверяет, что все обязательные поля заданы
func (cfg *VfsConfig) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("missing field: Endpoint")
	}
	if cfg.AccessKeyID == "" {
		return errors.New("missing field: AccessKeyID")
	}
	if cfg.SecretKey == "" {
		return errors.New("missing field: SecretKey")
	}
	if cfg.Bucket == "" {
		return errors.New("missing field: Bucket")
	}
	return nil
}

func NewVfs(cfg *VfsConfig) (Vfs, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Добавляем схему, если отсутствует
	if !strings.HasPrefix(cfg.Endpoint, "http://") && !strings.HasPrefix(cfg.Endpoint, "https://") {
		scheme := "http://"
		if cfg.UseSSL {
			scheme = "https://"
		}
		cfg.Endpoint = scheme + cfg.Endpoint
	}

	parsedUrl, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	// создаём minio.Client
	minioClient, err := minio.New(parsedUrl.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretKey, ""),
		Secure: parsedUrl.Scheme == "https",
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	location := s3_minio.NewLocation(minioClient)

	v := &vfsMinio{
		client:   minioClient,
		location: location,
		config:   cfg,
	}

	return v, nil
}

func (v *vfsMinio) Item(ctx context.Context, path string) (file s3_wrappers.Item, err error) {
	return v.getItem(path, v.config.Bucket)
}

func (v *vfsMinio) List(ctx context.Context, prefix string, pageSize int) (files []s3_wrappers.Item, err error) {
	err = v.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("error connect to filestorage. err: %s cfg: VfsKind: %s, VfsEndpoint: %s, VfsBucket: %s",
			err, v.config.Comma, v.config.Endpoint, v.config.Bucket)
	}
	defer v.Close()

	cursor := ""
	for {
		items, nextCursor, err := v.container.Items(ctx, prefix, cursor, pageSize)
		if err != nil {
			return nil, fmt.Errorf("error listing items from container: %w", err)
		}

		files = append(files, items...)

		if nextCursor == "" {
			break // больше нет страниц
		}
		cursor = nextCursor
	}

	return files, nil
}

func (v *vfsMinio) Read(ctx context.Context, file string, private_access bool) (data []byte, mimeType string, err error) {
	data, mimeType, err = v.ReadFromBucket(ctx, file, v.config.Bucket, private_access)
	if err != nil {
		if ctx.Err() != nil {
			return nil, "", fmt.Errorf("read failed due to context: %w", ctx.Err())
		}
		return nil, "", err
	}
	return data, mimeType, nil
}

func (v *vfsMinio) ReadFromBucket(ctx context.Context, file, bucket string, privateAccess bool) ([]byte, string, error) {
	if err := v.Connect(ctx); err != nil {
		return nil, "", fmt.Errorf("connect error: %w (endpoint: %s, bucket: %s)", err, v.config.Endpoint, v.config.Bucket)
	}
	defer v.Close()

	reader, err := v.ReadCloserFromBucket(ctx, file, bucket, privateAccess)
	if err != nil {
		return nil, "", fmt.Errorf("ReadCloserFromBucket failed: %w", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("error reading data: %w", err)
	}

	mimeType := detectMIME(data, file)
	return data, mimeType, nil
}

func (v *vfsMinio) ReadCloser(ctx context.Context, file string, private_access bool) (reader io.ReadCloser, err error) {
	return v.ReadCloserFromBucket(ctx, file, v.config.Bucket, private_access)
}

func (v *vfsMinio) ReadCloserFromBucket(ctx context.Context, file, bucket string, private_access bool) (reader io.ReadCloser, err error) {
	user, _ := ctx.Value(userUid).(string)

	if strings.Contains(file, "users") && (user == "" || !strings.Contains(file, user)) && !private_access {
		return nil, errors.New(privateDirectory)
	}

	item, err := v.getItem(file, bucket)
	if err != nil {
		return nil, err
	}

	reader, err = item.Open(ctx)
	if err != nil {
		return nil, err
	}

	return reader, err
}

func (v *vfsMinio) Write(ctx context.Context, file string, data []byte) (err error) {
	type result struct {
		I   s3_wrappers.Item
		Err error
	}

	err = v.Connect(ctx)
	if err != nil {
		return fmt.Errorf("error connect to filestorage. err: %s cfg: VfsKind: %s, VfsEndpoint: %s, VfsBucket: %s", err, minioVfsKind, v.config.Endpoint, v.config.Bucket)
	}
	defer v.Close()

	sdata := string(data)
	r := strings.NewReader(sdata)
	size := int64(len(sdata))

	// если передан разделитель, то заменяем / на него (возможно понадобится для совместимости плоских хранилищ)
	if v.config.Comma != "" {
		file = strings.Replace(file, sep, v.config.Comma, -1)
	}

	if strings.Contains(file, "../") {
		return fmt.Errorf("path file not valid")
	}

	chResult := make(chan result)
	exec := func(ctx context.Context, name string, rr io.Reader, size int64, metadata map[string]interface{}) (r result) {
		_, err = v.container.Put(ctx, file, rr, size, nil)
		r.Err = err
		return r
	}

	go func() {
		chResult <- exec(ctx, file, r, size, nil)
	}()

	select {
	case d := <-chResult:
		return d.Err
	case <-ctx.Done():
		return fmt.Errorf("exec Write dead for context")
	}
}

func (v *vfsMinio) Delete(ctx context.Context, file string) (err error) {
	err = v.Connect(ctx)
	if err != nil {
		return fmt.Errorf("error connect to filestorage. err: %s cfg: VfsKind: %s, VfsEndpoint: %s, VfsBucket: %s", err, minioVfsKind, v.config.Endpoint, v.config.Bucket)
	}
	defer v.Close()

	item, err := v.getItem(file, v.config.Bucket)
	if err != nil {
		return fmt.Errorf("error get Item for path: %s, err: %s", file, err)
	}

	err = v.container.RemoveItem(ctx, item.ID())
	if err != nil {
		return err
	}

	return err
}

func (v *vfsMinio) Connect(ctx context.Context) error {
	if v.location != nil && v.container != nil {
		return nil // уже подключено
	}

	// Подключаемся к MinIO
	loc := s3_minio.NewLocation(v.client)
	v.location = loc

	// Проверяем, существует ли контейнер
	container, err := loc.Container(ctx, v.config.Bucket)
	if err != nil {
		// Если бакет не найден — пробуем создать
		container, err = loc.CreateContainer(ctx, v.config.Bucket)
		if err != nil {
			return fmt.Errorf("failed to create container %q: %w", v.config.Bucket, err)
		}
	}
	v.container = container

	return nil
}

func (v *vfsMinio) Close() (err error) {
	err = v.location.Close()

	return err
}

func (v *vfsMinio) Proxy(trimPrefix, newPrefix string) (http.Handler, error) {
	// 1. Собираем URL целевого S3-совместимого эндпоинта с учётом схемы
	scheme := "https"
	if !v.config.UseSSL {
		scheme = "http"
	}

	// убедимся, что endpoint без схемы, иначе double-scheme
	endpoint := strings.TrimPrefix(v.config.Endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	parsedURL, err := url.Parse(fmt.Sprintf("%s://%s", scheme, endpoint))
	if err != nil {
		return nil, err
	}

	// 2. Создаём полноценный ReverseProxy с Rewrite
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			r.SetURL(parsedURL)

			// Преобразуем путь, если задан trimPrefix
			if trimPrefix != "" && strings.HasPrefix(r.Out.URL.Path, trimPrefix) {
				r.Out.URL.Path = newPrefix + strings.TrimPrefix(r.Out.URL.Path, trimPrefix)
			}
		},

		ModifyResponse: func(resp *http.Response) error {
			resp.Header.Del("Server")
			for k := range resp.Header {
				if strings.HasPrefix(k, "X-Amz-") {
					resp.Header.Del(k)
				}
			}

			if resp.StatusCode == http.StatusNotFound {
				resp.Body = io.NopCloser(bytes.NewReader(nil))
				resp.Header.Del("Content-Type")
				resp.Header.Set("Content-Length", "0")
				resp.ContentLength = 0
			}
			return nil
		},
	}

	// 3. Прокси будет использовать наш кастомный транспорт с SigV4
	t := &BasicAuthTransport{
		Kind:       minioVfsKind, // "s3"
		Username:   v.config.AccessKeyID,
		Password:   v.config.SecretKey,
		Region:     v.config.Region,
		DisableSSL: !v.config.UseSSL,
	}
	if v.config.CACert != "" {
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM([]byte(v.config.CACert))
		t.TLSClientConfig = &tls.Config{RootCAs: pool}
	}
	proxy.Transport = t

	return proxy, nil
}

func (v *vfsMinio) getItem(file, bucket string) (item s3_wrappers.Item, err error) {
	var urlPath url.URL

	// если передан разделитель, то заменяем / на него (возможно понадобится для совместимости плоских хранилищ)
	if v.config.Comma != "" {
		file = strings.Replace(file, v.config.Comma, sep, -1)
	}

	//// если локально, то добавляем к endpoint бакет
	//if v.config.Kind == "local" {
	//	file = v.endpoint + sep + bucket + sep + file
	//	// подчищаем //
	//	file = strings.Replace(file, sep+sep, sep, -1)
	//} else {
	// подчищаем от части путей, которая использовалась раньше в локальном хранилище
	// легаси, удалить когда все сайты переедут на использование только vfs
	//localPrefix := sep + "upload" + sep + v.bucket
	localPrefix := "upload" + sep + bucket
	file = strings.Replace(file, localPrefix, "", -1)
	file = strings.Replace(file, sep+sep, sep, -1)
	//}

	//fmt.Printf("file: %s, bucket: %s, container: %-v\n", file, bucket, v.container)
	urlPath.Path = "/" + bucket + "/" + strings.TrimPrefix(file, "/")

	if v.location == nil {
		return nil, fmt.Errorf("error. location is empty. bucket: %s, file: %s, endpoint: %s", urlPath.Host, urlPath.Path, v.config.Endpoint)
	}

	item, err = v.location.ItemByURL(&urlPath)
	if err != nil {
		return nil, fmt.Errorf("error. location.ItemByURL is failled. urlPath: %v, err: %s", urlPath, err)
	}

	if item == nil {
		return nil, fmt.Errorf("error. Item is null. urlPath: %v", urlPath)
	}

	return item, err
}
