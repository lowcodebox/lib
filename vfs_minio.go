// lib/vfs_minio.go
package lib

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib/internal/utils"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3_minio"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3_wrappers"
	"github.com/go-playground/validator/v10"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	minioVfsKind = "s3"
	maxExpiry    = 7 * 24 * time.Hour
)

type vfsMinio struct {
	validate   *validator.Validate
	baseClient *minio.Client
	cdnClient  *minio.Client
	location   s3_wrappers.Location
	container  s3_wrappers.Container
	config     *models.VFSConfig
}

//// VfsConfig — конфиг для MinIO/S3
//type VfsConfig struct {
//	Endpoint    string // host:port или полный URL
//	AccessKeyID string
//	SecretKey   string
//	Region      string
//	Bucket      string
//	UseSSL      bool   // https или http
//	Comma       string // заменитель для "/"
//	CACert      string // если нужен кастомный CA — можно передать, но в этом примере не используется
//}

// Validate проверяет, что все обязательные поля заданы
func ValidateVFSConfig(cfg *models.VFSConfig) error {
	if cfg.VfsEndpoint == "" {
		return errors.New("missing field: Endpoint")
	}
	if cfg.VfsAccessKeyID == "" {
		return errors.New("missing field: AccessKeyID")
	}
	if cfg.VfsSecretKey == "" {
		return errors.New("missing field: SecretKey")
	}
	if cfg.VfsBucket == "" {
		return errors.New("missing field: Bucket")
	}
	return nil
}

func NewVfs(cfg *models.VFSConfig) (Vfs, error) {
	if err := ValidateVFSConfig(cfg); err != nil {
		return nil, err
	}

	// Добавляем схему, если отсутствует
	if !strings.HasPrefix(cfg.VfsEndpoint, "http://") && !strings.HasPrefix(cfg.VfsEndpoint, "https://") {
		scheme := "http://"
		if cfg.VfsCertCA != "" || cfg.VfsCAFile != "" {
			scheme = "https://"
		}
		cfg.VfsEndpoint = scheme + cfg.VfsEndpoint
	}

	parsedUrl, err := url.Parse(cfg.VfsEndpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}

	var transport http.RoundTripper
	if cfg.VfsCertCA != "" || cfg.VfsCAFile != "" {
		// Создаём пул и добавляем кастомный CA

		if cfg.VfsCAFile != "" {
			caFileContent, err := os.ReadFile(cfg.VfsCAFile)
			if err != nil {
				return nil, err
			}
			cfg.VfsCertCA = string(caFileContent)
		}

		rootCAs := x509.NewCertPool()
		if ok := rootCAs.AppendCertsFromPEM([]byte(cfg.VfsCertCA)); !ok {
			return nil, fmt.Errorf("failed to append CA cert")
		}

		tlsConfig := &tls.Config{
			RootCAs:            rootCAs,
			InsecureSkipVerify: true,
		}

		transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	// создаём minio.Client
	baseMinioClient, err := minio.New(parsedUrl.Host, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.VfsAccessKeyID, cfg.VfsSecretKey, ""),
		Secure:    parsedUrl.Scheme == "https",
		Region:    cfg.VfsRegion,
		Transport: transport,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize minio client: %w", err)
	}

	var cdnMinioClient *minio.Client
	if isVfsCDNUserAvailable(cfg) {
		cdnMinioClient, err = minio.New(parsedUrl.Host, &minio.Options{
			Creds:     credentials.NewStaticV4(cfg.VfsCDNAccessKeyID, cfg.VfsCDNSecretKey, ""),
			Secure:    parsedUrl.Scheme == "https",
			Region:    cfg.VfsRegion,
			Transport: transport,
		})
		if err != nil {
			return nil, err
		}
	}

	location := s3_minio.NewLocation(baseMinioClient)

	v := &vfsMinio{
		baseClient: baseMinioClient,
		cdnClient:  cdnMinioClient,
		location:   location,
		config:     cfg,
		validate:   validator.New(),
	}

	return v, nil
}

func isVfsCDNUserAvailable(cfg *models.VFSConfig) bool {
	return cfg.VfsCDNAccessKeyID != "" && cfg.VfsCDNSecretKey != ""
}

func (v *vfsMinio) Item(ctx context.Context, path string) (file s3_wrappers.Item, err error) {
	return v.getItem(path, v.config.VfsBucket)
}

func (v *vfsMinio) List(ctx context.Context, prefix string, pageSize int) (files []s3_wrappers.Item, err error) {
	err = v.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("error connect to filestorage. err: %s cfg: VfsKind: %s, VfsEndpoint: %s, VfsBucket: %s",
			err, v.config.VfsComma, v.config.VfsEndpoint, v.config.VfsBucket)
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
	data, mimeType, err = v.ReadFromBucket(ctx, file, v.config.VfsBucket, private_access)
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
		return nil, "", fmt.Errorf("connect error: %w (endpoint: %s, bucket: %s)", err, v.config.VfsEndpoint, v.config.VfsBucket)
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
	return v.ReadCloserFromBucket(ctx, file, v.config.VfsBucket, private_access)
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
		return fmt.Errorf("error connect to filestorage. err: %s cfg: VfsKind: %s, VfsEndpoint: %s, VfsBucket: %s", err, minioVfsKind, v.config.VfsEndpoint, v.config.VfsBucket)
	}
	defer v.Close()

	sdata := string(data)
	r := strings.NewReader(sdata)
	size := int64(len(sdata))

	// если передан разделитель, то заменяем / на него (возможно понадобится для совместимости плоских хранилищ)
	if v.config.VfsComma != "" {
		file = strings.Replace(file, sep, v.config.VfsComma, -1)
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
		return fmt.Errorf("error connect to filestorage. err: %s cfg: VfsKind: %s, VfsEndpoint: %s, VfsBucket: %s", err, minioVfsKind, v.config.VfsEndpoint, v.config.VfsBucket)
	}
	defer v.Close()

	item, err := v.getItem(file, v.config.VfsBucket)
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
	loc := s3_minio.NewLocation(v.baseClient)
	v.location = loc

	// Проверяем, существует ли контейнер
	container, err := loc.Container(ctx, v.config.VfsBucket)
	if err != nil {
		//fmt.Printf("not found %s v.config.Bucket. err: %s", v.config.Bucket, err.Error())
		// Если бакет не найден — пробуем создать
		container, err = loc.CreateContainer(ctx, v.config.VfsBucket)
		if err != nil {
			return fmt.Errorf("failed to create container %q: %w", v.config.VfsBucket, err)
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Обработка пути
		trimmed := strings.TrimPrefix(r.URL.Path, trimPrefix)
		objectPath := utils.JoinURLPath(newPrefix, trimmed)

		decodedPath, err := url.PathUnescape(objectPath)
		if err != nil {
			http.Error(w, "invalid path", http.StatusBadRequest)
			return
		}

		parts := strings.SplitN(strings.TrimLeft(decodedPath, "/"), "/", 2)
		if len(parts) != 2 {
			http.Error(w, "invalid bucket/key path", http.StatusBadRequest)
			return
		}
		bucket := parts[0]
		objectKey := parts[1]

		// Получаем объект
		obj, err := v.baseClient.GetObject(r.Context(), bucket, objectKey, minio.GetObjectOptions{})
		if err != nil {
			http.Error(w, fmt.Sprintf("error fetching object: %v", err), http.StatusBadGateway)
			return
		}
		defer obj.Close()

		// Получаем метаданные
		stat, err := obj.Stat()
		if err != nil {
			http.Error(w, fmt.Sprintf("error stat object: %v", err), http.StatusBadGateway)
			return
		}

		filename := filepath.Base(objectKey)
		ext := strings.ToLower(filepath.Ext(filename))

		// Исключение для HTML/HTM
		if ext == ".html" || ext == ".htm" {
			// Можно вообще не ставить Content-Disposition,
			// либо поставить inline
			w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
		} else {
			// Остальные — как файл на скачивание
			disposition := fmt.Sprintf("attachment; filename=%q", url.PathEscape(filename))
			w.Header().Set("Content-Disposition", disposition)
		}

		// Создаём виртуальный io.ReadSeeker (требуется ServeContent)
		reader := &utils.ReadSeekWrapper{
			ReadSeeker: obj,
			Closer:     obj,
		}

		// Установка заголовков
		w.Header().Set("ETag", stat.ETag)

		// ServeContent — сам обработает Range, HEAD, If-Modified-Since и т.п.
		http.ServeContent(w, r, filepath.Base(objectKey), stat.LastModified, reader)
	}), nil
}

func (v *vfsMinio) GetPresignedURL(ctx context.Context, in *PresignedURLIn) (url string, err error) {
	if err := v.validateCDNClient(); err != nil {
		return "", err
	}
	if err := v.validate.Struct(in); err != nil {
		return "", err
	}

	// Ограничение максимального срока действия (совместимо с S3)
	expiry := in.Duration
	if expiry > maxExpiry {
		expiry = maxExpiry
	}

	object := strings.TrimPrefix(in.Path, "/")

	// Генерация presigned URL
	u, err := v.cdnClient.PresignedGetObject(ctx, in.Bucket, object, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to presign object: %w", err)
	}

	return u.String(), nil
}

func (v *vfsMinio) PutPresignedURL(ctx context.Context, in *PresignedURLIn) (url string, err error) {
	if err := v.validateCDNClient(); err != nil {
		return "", err
	}
	if err := v.validate.Struct(in); err != nil {
		return "", err
	}

	// Ограничение максимального срока действия (совместимо с S3)
	expiry := in.Duration
	if expiry > maxExpiry {
		expiry = maxExpiry
	}

	object := strings.TrimPrefix(in.Path, "/")

	// Генерация presigned URL
	u, err := v.cdnClient.PresignedPutObject(ctx, in.Bucket, object, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to presign object: %w", err)
	}

	return u.String(), nil
}

func (v *vfsMinio) FileExists(ctx context.Context, in *PresignedURLIn) (exists bool, err error) {

	_, err = v.baseClient.StatObject(ctx, in.Bucket, in.Path, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil // Объект не существует
		}
		return false, err // Другая ошибка
	}
	return true, nil // Объект существует
}

func (v *vfsMinio) validateCDNClient() error {
	if v.cdnClient == nil {
		return errors.New("CDN client is not available. CDN keys is not defined in config")
	}
	return nil
}

func (v *vfsMinio) getItem(file, bucket string) (item s3_wrappers.Item, err error) {
	var urlPath url.URL

	// если передан разделитель, то заменяем / на него (возможно понадобится для совместимости плоских хранилищ)
	if v.config.VfsComma != "" {
		file = strings.Replace(file, v.config.VfsComma, sep, -1)
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
		return nil, fmt.Errorf("error. location is empty. bucket: %s, file: %s, endpoint: %s", urlPath.Host, urlPath.Path, v.config.VfsEndpoint)
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
