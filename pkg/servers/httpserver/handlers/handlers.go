package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/app/pkg/service"
	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
)

type handlers struct {
	service service.Service
	logger  lib.Log
	cfg     model.Config
}

type Handlers interface {
	Alive(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	Page(w http.ResponseWriter, r *http.Request)
	Block(w http.ResponseWriter, r *http.Request)
	Cache(w http.ResponseWriter, r *http.Request)
	AuthChangeRole(w http.ResponseWriter, r *http.Request)
	Storage(w http.ResponseWriter, r *http.Request)
}

func (h *handlers) transportResponse(w http.ResponseWriter, response interface{}) (err error) {
	w.WriteHeader(200)
	d, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(403)
	}
	w.Write(d)
	return err
}

func (h *handlers) transportError(w http.ResponseWriter, code int, error error, message string) (err error) {
	var res = models.Response{}

	res.Status.Error = error
	res.Status.Description = message
	d, err := json.Marshal(res)

	h.logger.Error(err, message)

	w.WriteHeader(code)
	w.Write(d)
	return err
}

func (h *handlers) transportByte(w http.ResponseWriter, mimeType string, response []byte) (err error) {
	if mimeType != "" {
		w.Header().Set("content-type", mimeType)
		w.Header().Set("content-length", fmt.Sprint(len(response)))
		w.Header().Set("accept-ranges", "bytes")
	}
	w.WriteHeader(200)
	if err != nil {
		w.WriteHeader(403)
	}

	//fmt.Println("\n\n", w.Header().Values("content-type"))
	//fmt.Println("\n\n", w.Header().Values("content-length"))

	w.Write(response)
	return err
}

func (h *handlers) transportResponseHTTP(w http.ResponseWriter, response string) (err error) {
	w.WriteHeader(200)

	if err != nil {
		w.WriteHeader(403)
	}
	w.Write([]byte(response))
	return err
}

func New(
	service service.Service,
	logger lib.Log,
	cfg model.Config,
) Handlers {
	return &handlers{
		service,
		logger,
		cfg,
	}
}
