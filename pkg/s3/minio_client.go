package s3

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ClientMinioBuilder строит клиент MinIO по конфигу ConfigS3.
type ClientMinioBuilder struct {
	config                *ConfigS3
	httpTransport         http.RoundTripper
	options               *MinioOptionsExtended
	sessionNClientCreator SessionClientCreator[*minio.Client, *MinioOptionsExtended]
}

type MinioOptionsExtended struct {
	minio.Options
	Endpoint string
}

// NewClientMinioBuilder возвращает билдера с настройками по умолчанию.
func NewClientMinioBuilder() IClientS3Builder[*minio.Client, *MinioOptionsExtended] {
	return &ClientMinioBuilder{
		config:        nil,
		httpTransport: http.DefaultTransport,
		options:       &MinioOptionsExtended{},
		// По умолчанию создаём клиент через minio.New
		sessionNClientCreator: func(opts *MinioOptionsExtended) (s3Client *minio.Client, err error) {
			return minio.New(opts.Endpoint, &opts.Options)
		},
	}
}

// SetConfig сохраняет пользовательский ConfigS3.
func (b *ClientMinioBuilder) SetConfig(config *ConfigS3) IClientS3Builder[*minio.Client, *MinioOptionsExtended] {
	b.config = config
	return b
}

// SetSessionNClientCreator позволяет подменить способ создания MinIO-клиента (для моков, тестов и т.п.).
func (b *ClientMinioBuilder) SetSessionNClientCreator(
	creator SessionClientCreator[*minio.Client, *MinioOptionsExtended],
) IClientS3Builder[*minio.Client, *MinioOptionsExtended] {
	b.sessionNClientCreator = creator
	return b
}

// Build собираетMinioOptionsExtended на основе ConfigS3 и возвращает *minio.Client.
func (b *ClientMinioBuilder) Build(ctx context.Context) (*ClientS3Result[*minio.Client], error) {
	// Настраиваем опции
	b.options.Endpoint = b.config.Endpoint
	b.options.Creds = credentials.NewStaticV4(
		b.config.AccessKeyID,
		b.config.SecretKey,
		"",
	)
	b.options.Secure = !b.config.DisableSSL
	b.options.Region = b.config.Region
	b.options.Transport = b.httpTransport

	// Если передан свой CA-сертификат — ставим его в HTTP-транспорт
	if b.config.CACertPEM != "" {
		b.setCACert(b.config.CACertPEM)
	}

	// Создаём клиент
	client, err := b.sessionNClientCreator(b.options)
	if err != nil {
		return nil, err
	}

	return &ClientS3Result[*minio.Client]{
		Client:   client,
		Endpoint: b.config.Endpoint,
	}, nil
}

// setCACert добавляет в пул доверенных CA один сертификат из PEM.
func (b *ClientMinioBuilder) setCACert(caPEM string) IClientS3Builder[*minio.Client, *MinioOptionsExtended] {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM([]byte(caPEM))

	b.httpTransport = &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: pool,
		},
	}
	return b
}
