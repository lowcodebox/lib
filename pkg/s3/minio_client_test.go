package s3_test

import (
	"errors"
	"net/http"
	"testing"

	minio "github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3"
)

func TestClientMinioBuilder(t *testing.T) {
	tests := []struct {
		name       string
		config     s3.ConfigS3
		mockClient *minio.Client
		mockErr    error
		wantErr    bool
	}{
		{
			name: "success without CA",
			config: s3.ConfigS3{
				Region:   "r1",
				Endpoint: "e1",
				// DisableSSL и CACertPEM остаются zero-valued
			},
			mockClient: &minio.Client{},
		},
		{
			name: "success with CA",
			config: s3.ConfigS3{
				Region:     "r2",
				Endpoint:   "e2",
				DisableSSL: true,
				CACertPEM:  "invalid-pem", // любой непустой PEM заставит ветку CA сработать
			},
			mockClient: &minio.Client{},
		},
		{
			name:    "session error",
			config:  s3.ConfigS3{Region: "r3", Endpoint: "e3"},
			mockErr: errors.New("fail"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotEndpoint, gotRegion string
			var gotSecure bool
			var gotTransport http.RoundTripper

			builder := s3.NewClientMinioBuilder().
				SetConfig(&tt.config).
				SetSessionNClientCreator(func(opts *s3.MinioOptionsExtended) (*minio.Client, error) {
					gotEndpoint = opts.Endpoint
					gotRegion = opts.Region
					gotSecure = opts.Secure
					gotTransport = opts.Transport
					return tt.mockClient, tt.mockErr
				})

			res, err := builder.Build(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.mockClient, res.Client)
			assert.Equal(t, tt.config.Endpoint, res.Endpoint)

			// проверяем, что опции прокинулись правильно
			assert.Equal(t, tt.config.Endpoint, gotEndpoint)
			assert.Equal(t, tt.config.Region, gotRegion)
			assert.Equal(t, !tt.config.DisableSSL, gotSecure)

			// без CA должен использоваться default-транспорт
			if tt.config.CACertPEM == "" {
				assert.Equal(t, http.DefaultTransport, gotTransport)
			} else {
				// с CA — transport должен измениться
				assert.NotEqual(t, http.DefaultTransport, gotTransport)
				tr, ok := gotTransport.(*http.Transport)
				if !ok {
					t.Fatalf("expected *http.Transport, got %T", gotTransport)
				}
				// у TLSClientConfig обязательно есть поле RootCAs (даже если пустой пул)
				assert.NotNil(t, tr.TLSClientConfig.RootCAs)
			}
		})
	}
}

func TestBasicMinioClientDirector(t *testing.T) {
	config := &s3.ConfigS3{Region: "r", Endpoint: "e"}
	mockClient := &minio.Client{}

	builder := s3.NewClientMinioBuilder().
		SetSessionNClientCreator(func(opts *s3.MinioOptionsExtended) (*minio.Client, error) {
			return mockClient, nil
		})

	dir := s3.NewBasicS3ClientDirector(builder, config)

	res, err := dir.BuildS3Client(ctx)
	assert.NoError(t, err)
	assert.Equal(t, mockClient, res.Client)
	assert.Equal(t, config.Endpoint, res.Endpoint)
}

func TestMockMinioClientDirector(t *testing.T) {
	config := &s3.ConfigS3{Region: "r2", Endpoint: "e2"}

	cases := []struct {
		name    string
		mockErr error
		wantErr bool
	}{
		{"success", nil, false},
		{"failure", errors.New("oops"), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &minio.Client{}
			dir := s3.NewMockS3ClientDirector[*minio.Client](
				s3.NewClientMinioBuilder(),
				config,
				mockClient,
				tc.mockErr,
			)

			res, err := dir.BuildS3Client(ctx)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, mockClient, res.Client)
			assert.Equal(t, config.Endpoint, res.Endpoint)
		})
	}
}
