package lib

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
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

func escapeNonASCII(s string) string {
	var buf strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c > 127 {
			buf.WriteString(fmt.Sprintf("%%%02X", c))
		} else {
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, "../") {
		return nil, ErrPath
	}

	// Modify and encode the path
	modifiedPath := t.NewPrefix + strings.TrimPrefix(req.URL.Path, t.TrimPrefix)
	escapedPath := url.PathEscape(modifiedPath)
	req.URL.Path = modifiedPath
	req.URL.RawPath = escapedPath

	user, _ := req.Context().Value(userUid).(string)

	if strings.Contains(req.URL.Path, "users") && (user == "" || !strings.Contains(req.URL.Path, user)) {
		return nil, errors.New(privateDirectory)
	}

	if t.Username != "" {
		switch t.Kind {
		case "s3":
			fmt.Println("s3 Authorization")
			signer := v4.NewSigner(credentials.NewStaticCredentials(t.Username, t.Password, ""))
			_, err := signer.Sign(req, nil, "s3", t.Region, time.Now().UTC())
			if err != nil {
				return nil, err
			}
		default:
			fmt.Println("default Authorization")
			req.Header.Set("Authorization", fmt.Sprintf("Basic %s",
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
					t.Username, t.Password)))))
		}
	}

	h, _ := json.Marshal(req.Header)
	fmt.Printf("after headers: %s\n", string(h))

	return t.Transport.RoundTrip(req)
}

func (v *vfs) Proxy(trimPrefix, newPrefix string) (http.Handler, error) {
	trimPrefix = url.QueryEscape(trimPrefix)
	newPrefix = url.QueryEscape(newPrefix)
	parsedUrl, err := url.Parse(v.endpoint)
	if err != nil {
		return nil, err
	}

	proxy := httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			r.SetURL(parsedUrl)
		},
		ModifyResponse: func(resp *http.Response) error {
			resp.Header.Del("Server")
			for k := range resp.Header {
				if strings.HasPrefix(k, "X-Amz-") {
					resp.Header.Del(k)
				}
			}

			if resp.StatusCode == http.StatusNotFound {
				resp.Body = io.NopCloser(bytes.NewReader(nil))
				resp.Header.Del("Content-Type")
				resp.Header.Set("Content-Length", "0")
				resp.ContentLength = 0
			}

			return nil
		},
	}

	transport := BasicAuthTransport{
		Kind:       v.kind,
		Username:   v.accessKeyID,
		Password:   v.secretKey,
		TrimPrefix: trimPrefix,
		NewPrefix:  newPrefix,
		URL:        v.endpoint,
		Region:     v.region,
		DisableSSL: v.cacert == "",
	}

	if v.cacert != "" {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(v.cacert))

		transport.TLSClientConfig = &tls.Config{
			RootCAs: caCertPool,
		}
	}

	proxy.Transport = &transport

	return &proxy, nil
}
