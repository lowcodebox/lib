package s3

import (
	"context"
)

type SessionClientCreator[T, C any] func(awsConfig C) (s3Client T, err error)

type IClientS3Builder[T, C any] interface {
	SetConfig(config *ConfigS3) IClientS3Builder[T, C]
	SetSessionNClientCreator(creator SessionClientCreator[T, C]) IClientS3Builder[T, C]
	Build(ctx context.Context) (res *ClientS3Result[T], err error)
}

type ClientS3Result[T any] struct {
	Client   T
	Endpoint string
}

type ClientS3Director[T any] interface {
	BuildS3Client(ctx context.Context) (*ClientS3Result[T], error)
}

type BasicS3ClientDirector[T, C any] struct {
	builder IClientS3Builder[T, C]
	config  *ConfigS3
}

func NewBasicS3ClientDirector[T, C any](builder IClientS3Builder[T, C], config *ConfigS3) *BasicS3ClientDirector[T, C] {
	return &BasicS3ClientDirector[T, C]{
		builder: builder,
		config:  config,
	}
}

func (d *BasicS3ClientDirector[T, C]) BuildS3Client(ctx context.Context) (*ClientS3Result[T], error) {
	return d.builder.
		SetConfig(d.config).
		Build(ctx)
}

type MockS3ClientDirector[T, C any] struct {
	builder      IClientS3Builder[T, C]
	config       *ConfigS3
	mockClient   T
	mockBuildErr error
}

func NewMockS3ClientDirector[T, C any](builder IClientS3Builder[T, C], config *ConfigS3, mockClient T, mockBuildErr error) *MockS3ClientDirector[T, C] {
	return &MockS3ClientDirector[T, C]{
		builder:      builder,
		config:       config,
		mockClient:   mockClient,
		mockBuildErr: mockBuildErr,
	}
}

func (d *MockS3ClientDirector[T, C]) BuildS3Client(ctx context.Context) (*ClientS3Result[T], error) {
	return d.builder.
		SetConfig(d.config).
		SetSessionNClientCreator(func(awsConfig C) (s3Client T, err error) {
			return d.mockClient, d.mockBuildErr
		}).
		Build(ctx)
}
