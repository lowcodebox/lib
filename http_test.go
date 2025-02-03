package lib_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
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

	handler := lib.MiddlewareValidUri(key)(http.HandlerFunc(okHandler))
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
