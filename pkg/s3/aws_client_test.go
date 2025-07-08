package s3_test

import (
	"errors"
	"testing"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3"

	"github.com/aws/aws-sdk-go/aws"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

func TestClientAWSBuilder(t *testing.T) {
	tests := []struct {
		name       string
		config     s3.ConfigS3
		mockClient *awss3.S3
		mockErr    error
		wantErr    bool
	}{
		{
			name: "success without CA",
			config: s3.ConfigS3{
				Region:     "r1",
				Endpoint:   "e1",
				DisableSSL: false,
				CACertPEM:  "",
			},
			mockClient: &awss3.S3{},
		},
		{
			name: "success with CA",
			config: s3.ConfigS3{
				Region:     "r2",
				Endpoint:   "e2",
				DisableSSL: true,
				CACertPEM:  "invalid-pem", // just non-empty to hit CA branch
			},
			mockClient: &awss3.S3{},
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
			var gotRegion, gotEndpoint string
			builder := s3.NewClientAWSBuilder().
				SetConfig(&tt.config).
				SetSessionNClientCreator(func(cfg *aws.Config) (*awss3.S3, error) {
					gotRegion = aws.StringValue(cfg.Region)
					gotEndpoint = aws.StringValue(cfg.Endpoint)
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
			// verify the aws.Config was set up correctly
			assert.Equal(t, tt.config.Region, gotRegion)
			assert.Equal(t, tt.config.Endpoint, gotEndpoint)
		})
	}
}

func TestBasicAWSClientDirector(t *testing.T) {
	config := &s3.ConfigS3{Region: "r", Endpoint: "e"}
	mockClient := &awss3.S3{}
	builder := s3.NewClientAWSBuilder().
		SetSessionNClientCreator(func(_ *aws.Config) (*awss3.S3, error) {
			return mockClient, nil
		})
	dir := s3.NewBasicS3ClientDirector(builder, config)

	res, err := dir.BuildS3Client(ctx)
	assert.NoError(t, err)
	assert.Equal(t, mockClient, res.Client)
	assert.Equal(t, config.Endpoint, res.Endpoint)
}

func TestMockAWSClientDirector(t *testing.T) {
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
			mockClient := &awss3.S3{}
			dir := s3.NewMockS3ClientDirector[*awss3.S3](
				s3.NewClientAWSBuilder(),
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
