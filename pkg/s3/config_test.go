package s3_test

import (
	"context"
	"testing"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib/pkg/s3"
	"github.com/stretchr/testify/assert"
)

var ctx = context.Background()

func TestConfigS3Builder(t *testing.T) {
	tests := []struct {
		name            string
		initial         map[s3.ConfigField]string
		wantErr         bool
		wantAuthType    s3.AuthType
		wantAccessKeyID string
		wantSecretKey   string
		wantRegion      string
	}{
		{
			name: "accesskey minimal",
			initial: map[s3.ConfigField]string{
				s3.ConfigFieldAuthType:    "accesskey",
				s3.ConfigFieldAccessKeyID: "AKID123",
				s3.ConfigFieldSecretKey:   "SECRETXYZ",
				s3.ConfigFieldRegion:      "eu-west-1",
			},
			wantErr:         false,
			wantAuthType:    s3.AuthTypeAccessKey,
			wantAccessKeyID: "AKID123",
			wantSecretKey:   "SECRETXYZ",
			wantRegion:      "eu-west-1",
		},
		{
			name: "iam no keys",
			initial: map[s3.ConfigField]string{
				s3.ConfigFieldAuthType:    "iam",
				s3.ConfigFieldAccessKeyID: "",
				s3.ConfigFieldSecretKey:   "",
				s3.ConfigFieldRegion:      "",
			},
			wantErr:         false,
			wantAuthType:    s3.AuthTypeIAM,
			wantAccessKeyID: "",
			wantSecretKey:   "",
			wantRegion:      "us-east-1", // регион по умолчанию
		},
		{
			name: "missing key id",
			initial: map[s3.ConfigField]string{
				s3.ConfigFieldAuthType:  "accesskey",
				s3.ConfigFieldSecretKey: "SEC",
				s3.ConfigFieldRegion:    "eu-central-1",
				// access_key_id отсутствует
			},
			wantErr: true,
		},
		{
			name: "unsupported auth type",
			initial: map[s3.ConfigField]string{
				s3.ConfigFieldAuthType:    "google_auth",
				s3.ConfigFieldAccessKeyID: "A",
				s3.ConfigFieldSecretKey:   "B",
				s3.ConfigFieldRegion:      "eu-west-2",
			},
			wantErr: true,
		},
		{
			name: "accesskey empty key id",
			initial: map[s3.ConfigField]string{
				s3.ConfigFieldAuthType:    "accesskey",
				s3.ConfigFieldAccessKeyID: "",
				s3.ConfigFieldSecretKey:   "B",
				s3.ConfigFieldRegion:      "",
			},
			wantErr: true,
		},
		{
			name: "accesskey empty secret key",
			initial: map[s3.ConfigField]string{
				s3.ConfigFieldAuthType:    "accesskey",
				s3.ConfigFieldAccessKeyID: "A",
				s3.ConfigFieldSecretKey:   "",
				s3.ConfigFieldRegion:      "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// инициализируем "внешний" KVStore
			store := s3.NewLocalKVStore()
			err := s3.InitializeWithMap(ctx, store, tt.initial)
			assert.NoError(t, err, "Put should not fail for %s", tt.name)

			// DefaultConfigDirector внутри использует SetFieldsToUse(auth_type, access_key_id, secret_key, region)
			dir := s3.NewDefaultConfigDirector(s3.NewConfigS3Builder(), store)
			cfg, err := dir.BuildS3Config(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			// проверяем, что ошибки нет и поля совпадают
			assert.NoError(t, err)
			assert.Equal(t, tt.wantAuthType, cfg.AuthType)
			assert.Equal(t, tt.wantAccessKeyID, cfg.AccessKeyID)
			assert.Equal(t, tt.wantSecretKey, cfg.SecretKey)
			assert.Equal(t, tt.wantRegion, cfg.Region)
		})
	}
}
