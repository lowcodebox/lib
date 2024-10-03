package lib

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
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

func (t *BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, "../") {
		return nil, ErrPath
	}

	req.URL.Path = t.NewPrefix + strings.TrimPrefix(req.URL.Path, t.TrimPrefix)
	user, _ := req.Context().Value(userUid).(string)

	if strings.Contains(req.URL.Path, "users") && (user == "" || !strings.Contains(req.URL.Path, user)) {
		return nil, errors.New(privateDirectory)
	}
	if t.Username != "" {
		switch t.Kind {
		case "s3":
			//todo make sign for s3
			fmt.Println("s3 Authorization")
			signer := v4.NewSigner(credentials.NewStaticCredentials(t.Username, t.Password, ""))
			headers, err := signer.Sign(req, nil, t.URL, t.Region, time.Now().UTC())
			fmt.Printf("headers: %+v\n", headers)
			if err != nil {
				return nil, err
			}

			// fmt.Println(req.Header)

			// awsReq := request.Request{
			// 	Config: aws.Config{
			// 		CredentialsChainVerboseErrors: nil,
			// 		Credentials:                   credentials.NewStaticCredentials(t.Username, t.Password, ""),
			// 		Endpoint:                      aws.String(t.URL),
			// 		Region:                        aws.String(t.Region),
			// 		DisableSSL:                    aws.Bool(t.DisableSSL),
			// 		S3ForcePathStyle:              aws.Bool(true),
			// 	},
			// 	Time:        time.Now(),
			// 	HTTPRequest: req,
			// }

			// // Create a new SigV4 signer
			// signer := v4.NewSigner()

		default:
			fmt.Println("default Authorization")
			req.Header.Set("Authorization", fmt.Sprintf("Basic %s",
				base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
					t.Username, t.Password)))))
		}
	}
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		resp.Header.Set("Content-Type", "text/plain")
	}
	return resp, nil
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
