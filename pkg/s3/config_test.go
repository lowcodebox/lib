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
		initial         map[string]string
		wantErr         bool
		wantAuthType    s3.AuthType
		wantAccessKeyID string
		wantSecretKey   string
		wantRegion      string
	}{
		{
			name: "accesskey minimal",
			initial: map[string]string{
				"auth_type":     "accesskey",
				"access_key_id": "AKID123",
				"secret_key":    "SECRETXYZ",
				"region":        "eu-west-1",
			},
			wantErr:         false,
			wantAuthType:    s3.AuthTypeAccessKey,
			wantAccessKeyID: "AKID123",
			wantSecretKey:   "SECRETXYZ",
			wantRegion:      "eu-west-1",
		},
		{
			name: "iam no keys",
			initial: map[string]string{
				"auth_type":     "iam",
				"access_key_id": "",
				"secret_key":    "",
				"region":        "",
			},
			wantErr:         false,
			wantAuthType:    s3.AuthTypeIAM,
			wantAccessKeyID: "",
			wantSecretKey:   "",
			wantRegion:      "us-east-1", // регион по умолчанию
		},
		{
			name: "missing key id",
			initial: map[string]string{
				"auth_type":  "accesskey",
				"secret_key": "SEC",
				"region":     "eu-central-1",
				// access_key_id отсутствует
			},
			wantErr: true,
		},
		{
			name: "unsupported auth type",
			initial: map[string]string{
				"auth_type":     "google_auth",
				"access_key_id": "A",
				"secret_key":    "B",
				"region":        "eu-west-2",
			},
			wantErr: true,
		},
		{
			name: "accesskey empty key id",
			initial: map[string]string{
				"auth_type":     "accesskey",
				"access_key_id": "",
				"secret_key":    "B",
				"region":        "",
			},
			wantErr: true,
		},
		{
			name: "accesskey empty secret key",
			initial: map[string]string{
				"auth_type":     "accesskey",
				"access_key_id": "A",
				"secret_key":    "",
				"region":        "",
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
