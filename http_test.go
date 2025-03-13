package lib_test

import (
	"context"
	"encoding/json"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	key     = "0123456789abcdef"
	baseURL = "http://127.0.0.1"
)

var okResponse = models.Response{
	Status: models.RestStatus{
		Status: http.StatusOK,
	},
}

func okHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(okResponse)
	w.Write(resp)
}

func TestWhiteUri(t *testing.T) {
	type tcase struct {
		Path  string
		Token string
		Code  int
	}

	handler := lib.MiddlewareValidateUri(key)(http.HandlerFunc(okHandler))
	token, err := lib.NewServiceKey().WithWhiteURI("/metrics").Build([]byte(key))
	if err != nil {
		t.Fatal("Error generate service key", err)
	}

	tokenEmpty, err := lib.NewServiceKey().WithWhiteURI("").Build([]byte(key))
	if err != nil {
		t.Fatal("Error generate service key", err)
	}

	cases := []tcase{
		{Path: "/metrics", Token: token, Code: http.StatusOK},
		{Path: "/metrics?sec=1", Token: token, Code: http.StatusOK},
		{Path: "/metrics/jahlsd", Token: token, Code: http.StatusOK},
		{Path: "/metrics", Token: tokenEmpty, Code: http.StatusUnauthorized},
		{Path: "/metrics", Token: "lshjkdfgjbhkldsfgdgfhjkls", Code: http.StatusUnauthorized},
		{Path: "/metricsmonitor", Token: token, Code: http.StatusUnauthorized},
		{Path: "/notmetrics", Token: token, Code: http.StatusUnauthorized},
		{Path: "/lol", Token: token, Code: http.StatusUnauthorized},
	}

	for i, c := range cases {
		r := httptest.NewRequest(http.MethodGet, baseURL+c.Path, nil)
		w := httptest.NewRecorder()
		r.Header.Set(models.HeaderXServiceKey, c.Token)
		handler.ServeHTTP(w, r)

		if w.Code != c.Code {
			t.Errorf("Case %d. Expected code %d. Got %d", i, c.Code, w.Code)
		}
	}
}

func TestCurlV2(t *testing.T) {
	// Создаем тестовый сервер с редиректом
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			// Редирект на другой URL
			http.Redirect(w, r, "/after-redirect", http.StatusFound)
			return
		}

		if r.URL.Path == "/after-redirect" {
			// Возвращаем успешный ответ после редиректа
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Redirect successful"))
			return
		}

		// Возвращаем 404 для всех остальных запросов
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	ctx := context.Background()

	tests := []struct {
		name           string
		method         string
		url            string
		body           string
		response       interface{}
		headers        map[string]string
		cookies        []*http.Cookie
		enableRedirect bool
		timeout        time.Duration
		expStatusCode  int
	}{
		{
			name:           "not redirect",
			method:         http.MethodGet,
			url:            ts.URL + "/redirect",
			enableRedirect: true,
			expStatusCode:  http.StatusFound,
		},
		{
			name:           "after redirect",
			method:         http.MethodGet,
			url:            ts.URL + "/redirect",
			enableRedirect: false,
			expStatusCode:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, _, _, status, err := lib.CurlV2(
				ctx,
				tt.method,
				tt.url,
				tt.body,
				tt.response,
				tt.headers,
				tt.cookies,
				tt.timeout,
				tt.enableRedirect,
			)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if status != tt.expStatusCode {
				t.Fatalf("Expected status code %d. Got %d", tt.expStatusCode, status)
			}
		})
	}
}
