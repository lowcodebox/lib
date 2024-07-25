package curl

import (
	"crypto/tls"
	"net/http"
	"time"
)

const shortDuration = 1 * time.Second

var defaultRequestClient RequestClient

type RequestClient struct {
	client http.Client
}

func (r *RequestClient) NewRequest() *request {
	cloneRequest := defaultRequestClient.clone()

	return &request{
		client: cloneRequest.client,
	}
}

func (r *RequestClient) clone() RequestClient {
	return RequestClient{
		client: r.client,
	}
}

// RegisterDefaultClient регистрируем клиента по-умолчанию
func RegisterDefaultClient(timeout time.Duration) {
	defaultRequestClient = NewClient(timeout)

	return
}

// NewRequestDefault создаем запрос используя дефалтового клиента
func NewRequestDefault() *request {
	cloneRequest := defaultRequestClient.clone()

	return &request{
		client: cloneRequest.client,
	}
}

// NewClient создаем нового клиента
func NewClient(timeout time.Duration) RequestClient {
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	return RequestClient{
		client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: timeout,
		},
	}
}
