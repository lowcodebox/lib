package lib

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type BasicAuthTransport struct {
	Kind       string
	Username   string
	Password   string
	TrimPrefix string
	NewPrefix  string

	http.Transport
}

func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Path = t.NewPrefix + strings.TrimPrefix(req.URL.Path, t.TrimPrefix)

	if t.Username != "" {
		switch t.Kind {
		case "s3":
			// todo make sign for s3
			//	signer := v4.NewSigner()
			//	s3.Sign()

		default:
			req.Header.Set("Authorization", fmt.Sprintf("Basic %s",
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
					t.Username, t.Password)))))
		}
	}

	return t.Transport.RoundTrip(req)
}

func (v *vfs) Proxy(trimPrefix, newPrefix string) (http.Handler, error) {
	parsedUrl, err := url.Parse(v.endpoint)
	if err != nil {
		return nil, err
	}

	proxy := httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetXForwarded()
			r.SetURL(parsedUrl)
		},
	}

	transport := BasicAuthTransport{
		Kind:       v.kind,
		Username:   v.accessKeyID,
		Password:   v.secretKey,
		TrimPrefix: trimPrefix,
		NewPrefix:  newPrefix,
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
