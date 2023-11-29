// Package lib/vfs позволяет хранить файлы на разных источниках без необходимости учитывать особенности
// каждой реализации файлового хранилища
// поддерживаются local, s3, azure (остальные активировать по-необходимости)
package lib

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/lib/pkg/s3"
	"github.com/graymeta/stow"
	"github.com/graymeta/stow/azure"
	"github.com/graymeta/stow/local"

	// support Azure storage
	_ "github.com/graymeta/stow/azure"
	// support Google storage
	_ "github.com/graymeta/stow/google"
	// support local storage
	_ "github.com/graymeta/stow/local"
	// support swift storage
	_ "github.com/graymeta/stow/swift"
	// support s3 storage
	_ "github.com/graymeta/stow/s3"
	// support oracle storage
	_ "github.com/graymeta/stow/oracle"
)

type vfs struct {
	bucket                                         string
	kind, endpoint, accessKeyID, secretKey, region string
	location                                       stow.Location
	container                                      stow.Container
	comma                                          string
	cacert                                         string

	getTimeout, putTimeout, removeTimeout time.Duration
}

type Vfs interface {
	List(ctx context.Context, prefix string, pageSize int) (files []Item, err error)
	Read(ctx context.Context, file string) (data []byte, mimeType string, err error)
	ReadFromBucket(ctx context.Context, file, bucket string) (data []byte, mimeType string, err error)
	ReadCloser(ctx context.Context, file string) (reader io.ReadCloser, err error)
	ReadCloserFromBucket(ctx context.Context, file, bucket string) (reader io.ReadCloser, err error)
	Write(ctx context.Context, file string, data []byte) (err error)
	Delete(ctx context.Context, file string) (err error)
	Connect() (err error)
	Close() (err error)
}

type Item interface {
	stow.Item
}

// Connect инициируем подключение к хранилищу, в зависимости от типа соединения
func (v *vfs) Connect() (err error) {
	var config = stow.ConfigMap{}
	var flagBucketExist bool

	if v.region == "" {
		v.region = "eu-west-1"
	}
	switch v.kind {
	case "s3":
		config = stow.ConfigMap{
			s3.ConfigEndpoint:      v.endpoint,
			s3.ConfigAccessKeyID:   v.accessKeyID,
			s3.ConfigSecretKey:     v.secretKey,
			s3.ConfigRegion:        v.region,
			s3.ConfigCaCert:        v.cacert,
			s3.ConfigGetTimeout:    fmt.Sprint(v.getTimeout),
			s3.ConfigPutTimeout:    fmt.Sprint(v.putTimeout),
			s3.ConfigRemoveTimeout: fmt.Sprint(v.removeTimeout),
		}
	case "azure":
		config = stow.ConfigMap{
			azure.ConfigAccount: v.accessKeyID,
			azure.ConfigKey:     v.secretKey,
		}
	case "local":
		config = stow.ConfigMap{
			local.ConfigKeyPath: v.endpoint,
			local.MetadataDir:   v.bucket,
		}
	}

	// подсключаемся к хранилищу
	v.location, err = stow.Dial(v.kind, config)
	if err != nil {
		return fmt.Errorf("error create container from config. err: %s", err)
	}

	// ищем переданных бакет, если нет, то создаем его
	err = stow.WalkContainers(v.location, stow.NoPrefix, 10000, func(c stow.Container, err error) error {
		if err != nil {
			return err
		}
		if c.Name() == v.bucket {
			flagBucketExist = true
			return nil
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error list to containers from config. err: %s", err)
	}

	// создаем если нет
	if !flagBucketExist {
		v.container, err = v.location.CreateContainer(v.bucket)
		if err != nil {
			return fmt.Errorf("error create container from config. err: %s", err)
		}
	}

	// инициируем переданный контейнер
	v.container, err = v.location.Container(v.bucket)
	if err != nil {
		return fmt.Errorf("error create container from config. err: %s", err)
	}

	return err
}

// Close закрываем соединение
func (v *vfs) Close() (err error) {
	err = v.location.Close()

	return err
}

// Read чтение по указанному пути из бакета проекта
func (v *vfs) Read(ctx context.Context, file string) (data []byte, mimeType string, err error) {
	type result struct {
		Data     []byte
		MimeType string
		Err      error
	}

	chResult := make(chan result)
	exec := func(ctx context.Context, file string) (r result) {
		r.Data, r.MimeType, r.Err = v.ReadFromBucket(ctx, file, v.bucket)
		return r
	}
	go func() {
		chResult <- exec(ctx, file)
	}()

	select {
	case d := <-chResult:
		return d.Data, d.MimeType, d.Err

	case <-ctx.Done():
		return data, mimeType, fmt.Errorf("exec Read dead for context")
	}
}

// ReadFromBucket чтение по указанному пути из указанного бакета
func (v *vfs) ReadFromBucket(ctx context.Context, file, bucket string) (data []byte, mimeType string, err error) {
	type result struct {
		Reader io.ReadCloser
		Err    error
	}

	chResult := make(chan result)
	exec := func(ctx context.Context, file string) (r result) {
		r.Reader, r.Err = v.ReadCloserFromBucket(ctx, file, bucket)
		return r
	}
	go func() {
		chResult <- exec(ctx, file)
	}()

	select {
	case d := <-chResult:
		if d.Err != nil {
			return nil, "", err
		}
		data, err = ioutil.ReadAll(d.Reader)
		if err != nil {
			err = fmt.Errorf("error ReadAll. err: %s. file: %s, bucket: %s, v.container: %+v\n", err, file, bucket, v.container)
			return nil, "", err
		}
		mimeType = detectMIME(data, file) // - определяем MimeType отдаваемого файла

		return data, mimeType, err
	case <-ctx.Done():
		return data, mimeType, fmt.Errorf("exec ReadFromBucket dead for context")
	}
}

// Write создаем объект в хранилище
func (v *vfs) Write(ctx context.Context, file string, data []byte) (err error) {
	type result struct {
		I   Item
		Err error
	}

	sdata := string(data)
	r := strings.NewReader(sdata)
	size := int64(len(sdata))

	// если передан разделитель, то заменяем / на него (возможно понадобится для совместимости плоских хранилищ)
	if v.comma != "" {
		file = strings.Replace(file, sep, v.comma, -1)
	}

	chResult := make(chan result)
	exec := func(ctx context.Context, name string, rr io.Reader, size int64, metadata map[string]interface{}) (r result) {
		_, err = v.container.Put(file, rr, size, nil)
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

// Delete удаляем объект в хранилище
func (v *vfs) Delete(ctx context.Context, file string) (err error) {
	item, err := v.getItem(ctx, file, v.bucket)
	if err != nil {
		return fmt.Errorf("error get Item for path: %s, err: %s", file, err)
	}

	err = v.container.RemoveItem(item.ID())
	if err != nil {
		return err
	}
	return err
}

// List список файлов выбранного
func (v *vfs) List(ctx context.Context, prefix string, pageSize int) (files []Item, err error) {
	err = stow.Walk(v.container, prefix, pageSize, func(item stow.Item, err error) error {
		if err != nil {
			fmt.Printf("error Walk from list vfs. connect:%+v, prefix: %s, err: %s\n", v, prefix, err)
			return err
		}
		files = append(files, item)
		return nil
	})

	return files, err
}

func (v *vfs) ReadCloser(ctx context.Context, file string) (reader io.ReadCloser, err error) {
	return v.ReadCloserFromBucket(ctx, file, v.bucket)
}

func (v *vfs) ReadCloserFromBucket(ctx context.Context, file, bucket string) (reader io.ReadCloser, err error) {

	item, err := v.getItem(ctx, file, bucket)
	if err != nil {
		return nil, err
	}
	item.Open()

	reader, err = item.Open()

	return reader, err
}

func (v *vfs) getItem(ctx context.Context, file, bucket string) (item Item, err error) {
	var urlPath url.URL

	// если передан разделитель, то заменяем / на него (возможно понадобится для совместимости плоских хранилищ)
	if v.comma != "" {
		file = strings.Replace(file, v.comma, sep, -1)
	}

	// если локально, то добавляем к endpoint бакет
	if v.kind == "local" {
		file = v.endpoint + sep + bucket + sep + file
		// подчищаем //
		file = strings.Replace(file, sep+sep, sep, -1)
	} else {
		// подчищаем от части путей, которая использовалась раньше в локальном хранилище
		// легаси, удалить когда все сайты переедут на использование только vfs
		//localPrefix := sep + "upload" + sep + v.bucket
		localPrefix := "upload" + sep + bucket
		file = strings.Replace(file, localPrefix, "", -1)
		file = strings.Replace(file, sep+sep, sep, -1)
	}

	urlPath.Host = bucket
	urlPath.Path = file

	item, err = v.location.ItemByURL(&urlPath)
	if err != nil {
		return nil, fmt.Errorf("error. location.ItemByURL is failled. urlPath: %s, err: %s", urlPath, err)
	}
	if item == nil {
		return nil, fmt.Errorf("error. Item is null. urlPath: %s", urlPath)
	}

	return item, err
}

func NewVfs(kind, endpoint, accessKeyID, secretKey, region, bucket, comma, cacert string, s3GetTimeout, s3PutTimeout, s3RemoveTimeout time.Duration) Vfs {
	return &vfs{
		kind:        kind,
		endpoint:    endpoint,
		accessKeyID: accessKeyID,
		secretKey:   secretKey,
		region:      region,
		bucket:      bucket,
		comma:       comma,
		cacert:      cacert,

		getTimeout:    s3GetTimeout,
		putTimeout:    s3PutTimeout,
		removeTimeout: s3RemoveTimeout,
	}
}
