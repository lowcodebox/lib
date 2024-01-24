package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/app/pkg/service"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

type handlers struct {
	ctx     context.Context
	service service.Service
	cfg     model.Config
}

type Handlers interface {
	Alive(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	Page(w http.ResponseWriter, r *http.Request)
	Block(w http.ResponseWriter, r *http.Request)
	Cache(w http.ResponseWriter, r *http.Request)
	AuthChangeRole(w http.ResponseWriter, r *http.Request)
	AuthLogOut(w http.ResponseWriter, r *http.Request)
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

func (h *handlers) transportError(ctx context.Context, w http.ResponseWriter, code int, error error, message string) (err error) {
	var res = models.Response{}
	logger.Error(ctx, message, zap.Error(err))

	res.Status.Error = error
	res.Status.Description = message
	d, err := json.Marshal(res)
	if err != nil {
		return fmt.Errorf("error exec transportError. err: %s", err)
	}

	w.WriteHeader(code)
	_, err = w.Write(d)
	if err != nil {
		return fmt.Errorf("error exec transportError. err: %s", err)
	}

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

	_, err = w.Write(response)
	if err != nil {
		return fmt.Errorf("error exec transportError. err: %s", err)
	}

	return err
}

func (h *handlers) transportReader(w http.ResponseWriter, mimeType string, reader io.ReadCloser) (err error) {
	if mimeType != "" {
		w.Header().Set("content-type", mimeType)
		w.Header().Set("accept-ranges", "bytes")
	}
	w.WriteHeader(200)
	if err != nil {
		w.WriteHeader(403)
	}

	var ch chan []byte
	go func() {
		d := bufio.NewReader(reader)
		var buf = make([]byte, 1024)
		for {
			_, err = d.Read(buf)
			if err != nil {
				close(ch)
				return
			}
			ch <- buf
		}
	}()

	for {
		select {
		case d, ok := <-ch:
			_, err = w.Write(d)
			if err != nil {
				return err
			}
			if !ok {
				return err
			}
		}
	}
}

func (h *handlers) transportResponseHTTP(w http.ResponseWriter, response string) (err error) {
	w.WriteHeader(200)

	if err != nil {
		w.WriteHeader(403)
	}

	_, err = w.Write([]byte(response))
	if err != nil {
		return fmt.Errorf("error exec transportError. err: %s", err)
	}

	return err
}

func New(
	service service.Service,
	cfg model.Config,
) Handlers {
	ctx := context.Background()

	return &handlers{
		ctx,
		service,
		cfg,
	}
}
