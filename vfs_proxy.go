package lib

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

var ErrPath = errors.New("invalid path")

type BasicAuthTransport struct {
	Kind       string
	Username   string
	Password   string
	TrimPrefix string
	NewPrefix  string
	URL        string
	Region     string
	DisableSSL bool

	http.Transport
}

func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, "../") {
		return nil, ErrPath
	}

	user, _ := req.Context().Value(userUid).(string)

	if strings.Contains(req.URL.Path, "users") && (user == "" || !strings.Contains(req.URL.Path, user)) {
		return nil, errors.New(privateDirectory)
	}
	if t.Username != "" {
		switch t.Kind {
		case "s3":
			// Читаем тело
			var body []byte
			if req.Body != nil {
				b, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				body = b
				req.Body = io.NopCloser(bytes.NewReader(body))
			}

			// Вычисляем хеш тела (payload hash)
			var payloadHash string
			sum := sha256.Sum256(body)
			payloadHash = hex.EncodeToString(sum[:])

			// Устанавливаем host (важно для подписи)
			if req.URL.Host != "" {
				req.Host = req.URL.Host
			}

			// Устанавливаем Content-Length вручную
			if len(body) > 0 {
				req.ContentLength = int64(len(body))
			}

			// Подписываем запрос через SigV4
			signer := v4.NewSigner()
			creds := aws.Credentials{
				AccessKeyID:     t.Username,
				SecretAccessKey: t.Password,
			}
			err := signer.SignHTTP(
				context.Background(),
				creds,
				req,
				payloadHash,
				"s3",
				t.Region,
				time.Now(),
			)
			if err != nil {
				return nil, fmt.Errorf("sigv4 signing failed: %w", err)
			}

			// Восстанавливаем тело
			req.Body = io.NopCloser(bytes.NewReader(body))
		default:
			req.Header.Set("Authorization", fmt.Sprintf("Basic %s",
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
					t.Username, t.Password)))))
		}
	}

	return t.Transport.RoundTrip(req)
}
