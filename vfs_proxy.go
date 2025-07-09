package lib

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"io"
	"net/http"
	"time"
)

type BasicAuthTransport struct {
	Kind       string
	Username   string
	Password   string
	TrimPrefix string
	NewPrefix  string
	URL        string
	Region     string
	DisableSSL bool

	TLSClientConfig *tls.Config
}

func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// SigV4-ветка
	if t.Kind == "s3" {
		// читаем тело
		var body []byte
		if req.Body != nil {
			b, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			body = b
			req.Body = io.NopCloser(bytes.NewReader(body))
		}

		// готовим хэш
		payloadHash := "UNSIGNED-PAYLOAD"
		if len(body) > 0 {
			sum := sha256.Sum256(body)
			payloadHash = hex.EncodeToString(sum[:])
		}

		// подписываем
		signer := v4.NewSigner()
		creds := aws.Credentials{
			AccessKeyID:     t.Username,
			SecretAccessKey: t.Password,
		}
		if err := signer.SignHTTP(
			context.Background(),
			creds,
			req,
			payloadHash,
			"s3",
			t.Region,
			time.Now(),
		); err != nil {
			return nil, fmt.Errorf("sigv4 signing failed: %w", err)
		}

		// восстанавливаем Body для отправки
		req.Body = io.NopCloser(bytes.NewReader(body))

	} else if t.Username != "" {
		// Basic-авторизация (на всякий случай)
		req.Header.Set("Authorization", "Basic "+
			base64.StdEncoding.EncodeToString([]byte(t.Username+":"+t.Password)))
	}

	// по умолчанию — дефолтный транспорт (с нашим TLSConfig)
	tr := &http.Transport{TLSClientConfig: t.TLSClientConfig}
	return tr.RoundTrip(req)
}
