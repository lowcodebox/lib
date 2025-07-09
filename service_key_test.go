package lib_test

import (
	"testing"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"github.com/stretchr/testify/assert"
)

func TestXServiceKeyBuilder_FullFields_DefaultInterval(t *testing.T) {
	projectKey := []byte("test-project-key")
	now := time.Now()

	builder := lib.NewServiceKey().
		WithDomain("example.com").
		WithClient("client123").
		WithCheckCert(true).
		WithWhiteURI("/healthz").
		WithRole("admin").
		WithProfile("gold").
		WithGroups("g1,g2").
		WithServiceUID("svc-uid").
		WithRequestID("req-xyz")

	token, err := builder.Build(projectKey)
	assert.NoError(t, err, "Build should not return an error")
	assert.NotEmpty(t, token, "token should not be empty")

	out, err := lib.DecodeServiceKey(token, projectKey)
	assert.NoError(t, err, "DecodeServiceKey should succeed")

	// Check all fields were round-tripped
	assert.Equal(t, "example.com", out.Domain)
	assert.Equal(t, "client123", out.Client)
	assert.True(t, out.CheckCert)
	assert.Equal(t, "/healthz", out.WhiteURI)
	assert.Equal(t, "admin", out.Role)
	assert.Equal(t, "gold", out.Profile)
	assert.Equal(t, "g1,g2", out.Groups)
	assert.Equal(t, "svc-uid", out.ServiceUID)
	assert.Equal(t, "req-xyz", out.RequestID)

	// Expiration should be defaultTokenInterval = 5m (300s), within a few seconds tolerance
	expiredTime := time.Unix(out.Expired, 0)
	diff := expiredTime.Sub(now).Seconds()
	assert.InDelta(t, 300.0, diff, 5.0, "Expired should be ~5m after now")
}

// TODO - не стабильно работает
func TestXServiceKeyBuilder_CustomInterval_MinimalFields(t *testing.T) {
	t.Skip()
	projectKey := []byte("another-key")
	interval := 2 * time.Second
	nowUnix := time.Now().Unix()

	builder := lib.NewServiceKey().
		WithInterval(interval)

	token, err := builder.Build(projectKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	out, err := lib.DecodeServiceKey(token, projectKey)
	assert.NoError(t, err)

	// Все строковые поля пусты, булевы — false
	assert.Empty(t, out.Domain)
	assert.Empty(t, out.Client)
	assert.False(t, out.CheckCert)
	assert.Empty(t, out.WhiteURI)
	assert.Empty(t, out.Role)
	assert.Empty(t, out.Profile)
	assert.Empty(t, out.Groups)
	assert.Empty(t, out.ServiceUID)
	assert.Empty(t, out.RequestID)

	// Проверяем Expired: он должен быть > nowUnix и ≤ nowUnix + interval + 1s
	exp := out.Expired
	assert.Greater(t, exp, nowUnix, "Expired должен быть позже текущего времени")
	maxCustom := nowUnix + int64(interval.Seconds()) + 1
	assert.LessOrEqual(t, exp, maxCustom, "Expired не должен быть позже чем now+interval+1s")
}

func TestDecodeServiceKey_InvalidToken(t *testing.T) {
	_, err := lib.DecodeServiceKey("not-a-valid-token", []byte("key"))
	assert.Error(t, err, "Invalid token should produce an error")
}
