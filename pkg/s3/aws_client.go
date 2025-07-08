package s3

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"net/http"
	"time"
)

type ClientAWSBuilder struct {
	config                *ConfigS3
	httpClient            *http.Client
	awsConfig             *aws.Config
	sessionNClientCreator SessionClientCreator[*s3.S3, *aws.Config]
}

func NewClientAWSBuilder() IClientS3Builder[*s3.S3, *aws.Config] {
	return &ClientAWSBuilder{
		config:     nil,
		httpClient: http.DefaultClient,
		awsConfig:  aws.NewConfig(),
		sessionNClientCreator: func(awsConfig *aws.Config) (s3Client *s3.S3, err error) {
			sess, err := session.NewSession(awsConfig)
			if err != nil {
				return nil, err
			}
			if sess == nil {
				return nil, errors.New("creating the S3 session")
			}
			s3Client = s3.New(sess)
			return s3Client, nil
		},
	}
}

func (c *ClientAWSBuilder) SetConfig(config *ConfigS3) IClientS3Builder[*s3.S3, *aws.Config] {
	c.config = config
	return c
}

func (c *ClientAWSBuilder) SetSessionNClientCreator(creator SessionClientCreator[*s3.S3, *aws.Config]) IClientS3Builder[*s3.S3, *aws.Config] {
	c.sessionNClientCreator = creator
	return c
}

func (c *ClientAWSBuilder) Build(ctx context.Context) (res *ClientS3Result[*s3.S3], err error) {
	c.awsConfig.WithHTTPClient(c.httpClient).
		WithMaxRetries(aws.UseServiceDefaultRetries).
		WithLogger(aws.NewDefaultLogger()).
		WithLogLevel(aws.LogOff).
		WithSleepDelay(time.Sleep).
		WithRegion(c.config.Region).
		WithEndpoint(c.config.Endpoint).
		WithS3ForcePathStyle(true).
		WithDisableSSL(c.config.DisableSSL)

	if c.config.CACertPEM != "" {
		c.setCACert(c.config.CACertPEM)
	}

	s3Client, err := c.sessionNClientCreator(c.awsConfig)
	if err != nil {
		return nil, err
	}
	return &ClientS3Result[*s3.S3]{
		Client:   s3Client,
		Endpoint: c.config.Endpoint,
	}, nil
}

func (c *ClientAWSBuilder) setCACert(caCert string) IClientS3Builder[*s3.S3, *aws.Config] {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(caCert))

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	c.httpClient.Transport = transport
	return c
}
