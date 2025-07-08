package s3

import (
	"context"
	"fmt"
)

type AuthType string

const (
	AuthTypeAccessKey AuthType = "accesskey"
	AuthTypeIAM       AuthType = "iam"
)

type ConfigField string

const (
	ConfigFieldAuthType    ConfigField = "auth_type"
	ConfigFieldAccessKeyID ConfigField = "access_key_id"
	ConfigFieldSecretKey   ConfigField = "secret_key"
	ConfigFieldRegion      ConfigField = "region"
	ConfigFieldEndpoint    ConfigField = "endpoint"
	ConfigFieldDisableSSL  ConfigField = "disable_ssl"
	ConfigFieldV2Signing   ConfigField = "v2_signing"
	ConfigFieldCACert      ConfigField = "ca_cert"
)

var (
	allowedConfigFields = map[ConfigField]struct{}{
		ConfigFieldAuthType:    {},
		ConfigFieldAccessKeyID: {},
		ConfigFieldSecretKey:   {},
		ConfigFieldRegion:      {},
		ConfigFieldEndpoint:    {},
		ConfigFieldDisableSSL:  {},
		ConfigFieldV2Signing:   {},
		ConfigFieldCACert:      {},
	}
)

type IConfigS3Builder interface {
	SetKV(store KVStore) IConfigS3Builder
	SetFieldsToUse(fields []ConfigField) IConfigS3Builder
	Refill(key, val string) IConfigS3Builder
	Build(ctx context.Context) (config *ConfigS3, err error)
}

type ConfigS3Builder struct {
	fields       []ConfigField
	kvStore      KVStore
	localKVStore LocalKVStore
}

func (c *ConfigS3Builder) SetKV(store KVStore) IConfigS3Builder {
	c.kvStore = store
	return c
}

func (c *ConfigS3Builder) SetFieldsToUse(fields []ConfigField) IConfigS3Builder {
	c.fields = fields
	return c
}

func (c *ConfigS3Builder) Refill(key, val string) IConfigS3Builder {
	_ = c.localKVStore.Put(context.TODO(), key, val)
	return c
}
func (c *ConfigS3Builder) Build(ctx context.Context) (*ConfigS3, error) {
	// Валидация: только разрешённые поля
	for _, field := range c.fields {
		if _, ok := allowedConfigFields[field]; !ok {
			return nil, fmt.Errorf("field [%s] not allowed", field)
		}
	}

	localVals := make(map[ConfigField]string)

	// Сначала берём из localKVStore
	for _, field := range c.fields {
		ok, err := c.localKVStore.Check(ctx, string(field))
		if err != nil {
			return nil, err
		}
		if ok {
			val, err := c.localKVStore.Get(ctx, string(field))
			if err != nil {
				return nil, err
			}
			localVals[field] = val
		}
	}

	// Потом добираем из внешнего KVStore, если не найдено локально
	for _, field := range c.fields {
		if _, found := localVals[field]; found {
			continue
		}
		val, err := c.kvStore.Get(ctx, string(field))
		if err != nil {
			return nil, err
		}
		localVals[field] = val
	}

	// Собираем финальный ConfigS3
	cfg := &ConfigS3{
		AuthType:     AuthType(localVals["auth_type"]),
		AccessKeyID:  localVals["access_key_id"],
		SecretKey:    localVals["secret_key"],
		Region:       localVals["region"],
		Endpoint:     localVals["endpoint"],
		CACertPEM:    localVals["ca_cert"],
		DisableSSL:   localVals["disable_ssl"] == "true",
		UseV2Signing: localVals["v2_signing"] == "true",
	}

	// Валидация обязательных полей
	if cfg.AuthType == "" {
		cfg.AuthType = AuthTypeAccessKey
	}
	if cfg.AuthType != AuthTypeAccessKey && cfg.AuthType != AuthTypeIAM {
		return nil, fmt.Errorf("unsupported auth_type: %s", cfg.AuthType)
	}
	if cfg.AuthType == AuthTypeAccessKey {
		if cfg.AccessKeyID == "" {
			return nil, fmt.Errorf("missing required field: access_key_id")
		}
		if cfg.SecretKey == "" {
			return nil, fmt.Errorf("missing required field: secret_key")
		}
	}
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	return cfg, nil
}

func NewConfigS3Builder() IConfigS3Builder {
	return &ConfigS3Builder{}
}

type ConfigS3 struct {
	AuthType     AuthType
	AccessKeyID  string
	SecretKey    string
	Region       string
	Endpoint     string
	DisableSSL   bool
	UseV2Signing bool
	CACertPEM    string
}

type ConfigDirector interface {
	BuildS3Config(ctx context.Context) (*ConfigS3, error)
}
type DefaultConfigDirector struct {
	builder IConfigS3Builder
	store   KVStore
}

func NewDefaultConfigDirector(configBuilder IConfigS3Builder, store KVStore) ConfigDirector {
	return &DefaultConfigDirector{
		builder: configBuilder,
		store:   store,
	}
}

func (d *DefaultConfigDirector) BuildS3Config(ctx context.Context) (*ConfigS3, error) {
	return d.builder.
		SetKV(d.store).
		SetFieldsToUse([]ConfigField{
			ConfigFieldAuthType,
			ConfigFieldAccessKeyID,
			ConfigFieldSecretKey,
			ConfigFieldRegion,
		}).
		Build(ctx)
}
