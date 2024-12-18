package lib

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
)

// Curl всегде возвращает результат в интерфейс + ошибка (полезно для внешних запросов с неизвестной структурой)
// сериализуем в объект, при передаче ссылки на переменную типа
func Curl(ctx context.Context, method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, status int, err error) {
	r, _, status, err := curl_engine(ctx, method, urlc, bodyJSON, response, headers, cookies)
	return r, status, err
}

func CurlCookies(ctx context.Context, method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, resp_cookies []*http.Cookie, status int, err error) {
	return curl_engine(ctx, method, urlc, bodyJSON, response, headers, cookies)
}

func curl_engine(ctx context.Context, method, urlc, bodyJSON string, response interface{}, headers map[string]string, cookies []*http.Cookie) (result interface{}, resp_cookies []*http.Cookie, status int, err error) {
	// Create a custom transport with timeouts
	transport := createTransport()

	// Create HTTP client with the transport
	client := &http.Client{
		Timeout:   clientHttpTimeout,
		Transport: transport,
	}

	// Prepare request
	req, err := prepareRequest(ctx, method, urlc, bodyJSON)
	if err != nil {
		return nil, nil, status, fmt.Errorf("error preparing request: %w", err)
	}

	// Add headers and cookies
	addHeadersAndCookies(ctx, req, headers, cookies)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, http.StatusBadRequest, err
	}
	defer resp.Body.Close()

	// Read and process response
	return processResponse(resp, response)
}

func createTransport() *http.Transport {
	dialer := net.Dialer{
		Timeout: clientHttpTimeout,
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	return &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: clientHttpTimeout / 5,
		TLSClientConfig:     tlsConfig,
	}
}

func addHeadersAndCookies(ctx context.Context, req *http.Request, headers map[string]string, cookies []*http.Cookie) {
	// Add headers from context and custom headers
	httpClientHeaders(ctx, req, headers)

	// Add cookies if provided
	if cookies != nil {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}
}

// httpClientHeaders устанавливает заголовки реквеста из контекста и headers
func httpClientHeaders(ctx context.Context, req *http.Request, headers map[string]string) {
	if req == nil {
		return
	}
	for ctxField, headerField := range models.ProxiedHeaders {
		if value := getFieldCtx(ctx, ctxField); value != "" {
			req.Header.Add(headerField, value)
		}
	}

	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
}

func prepareRequest(ctx context.Context, method, urlc, bodyJSON string) (*http.Request, error) {
	if method == "" {
		method = "POST"
	}
	method = strings.TrimSpace(method)

	// Handle special cases JSONTOGET and JSONTOPOST
	switch {
	case method == "JSONTOGET" && bodyJSON != "" && strings.Contains(urlc, "?"):
		return prepareJSONToGETRequest(urlc, bodyJSON)
	case method == "JSONTOPOST" && bodyJSON != "":
		return prepareJSONToPOSTRequest(urlc, bodyJSON)
	default:
		return http.NewRequest(method, urlc, strings.NewReader(bodyJSON))
	}
}

func prepareJSONToGETRequest(urlc, bodyJSON string) (*http.Request, error) {
	var mapValues map[string]string
	values := url.Values{}

	err := json.Unmarshal([]byte(bodyJSON), &mapValues)
	if err != nil {
		return nil, fmt.Errorf("error Unmarshal in Curl, bodyJSON: %s, err: %s", bodyJSON, err)
	}

	for k, v := range mapValues {
		values.Set(k, v)
	}

	uri, _ := url.Parse(urlc)
	uri.RawQuery = values.Encode()
	urlc = uri.String()

	return http.NewRequest("GET", urlc, strings.NewReader(bodyJSON))
}

func prepareJSONToPOSTRequest(urlc, bodyJSON string) (*http.Request, error) {
	var mapValues map[string]string
	values := url.Values{}

	err := json.Unmarshal([]byte(bodyJSON), &mapValues)
	if err != nil {
		return nil, fmt.Errorf("error Unmarshal in Curl, bodyJSON: %s, err: %s", bodyJSON, err)
	}

	for k, v := range mapValues {
		values.Set(k, v)
	}

	req, err := http.NewRequest("POST", urlc, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.PostForm = values
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	return req, nil
}

func processResponse(resp *http.Response, response interface{}) (interface{}, []*http.Cookie, int, error) {
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", nil, resp.StatusCode, err
	}

	responseString := string(responseData)

	if response != nil {
		json.Unmarshal([]byte(responseString), &response)
	}

	return responseString, resp.Cookies(), resp.StatusCode, nil
}
