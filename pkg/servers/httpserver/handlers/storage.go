package handlers

import (
	"context"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"net/http"
	"strings"
)

func (h *handlers) Storage(w http.ResponseWriter, r *http.Request) {
	in, err := storageDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[Alive] Error function execution (storageDecodeRequest).")
		return
	}
	serviceResult, err := h.service.Storage(r.Context(), in)
	if err != nil {
		h.logger.Error(err, "[Alive] Error service execution (service.Alive).")
		h.transportError(w, 500, err, "[Alive] Error service execution (service.Alive).")
		return
	}
	response, _ := storageEncodeResponse(r.Context(), serviceResult)
	if err != nil {
		h.transportError(w, 500, err, "[Alive] Error function execution (storageEncodeResponse)")
		h.logger.Error(err, "[Alive] Error function execution (storageEncodeResponse).")
		return
	}

	err = h.transportByte(w, response.MimeType, response.Body)
	if err != nil {
		h.transportError(w, 500, err, "[Query] Error function execution (transportResponse)")
		return
	}

	return
}

func storageDecodeRequest(ctx context.Context, r *http.Request) (request model.StorageIn, err error) {
	// отрезаем первый раздел в пути, это /upload, путь к файлу дальше
	file := r.URL.Path
	fileName := strings.Split(r.URL.Path, "/")
	if len(fileName) > 1 {
		file = strings.Join(fileName[2:], "/")
	}

	request.Bucket = fileName[1]
	request.File = file

	return request, err
}

func storageEncodeResponse(ctx context.Context, serviceResult model.StorageOut) (response model.StorageOut, err error) {
	return serviceResult, err
}

func detectMimeType(r *http.Request) (mimeType string, err error) {

	return mimeType, err
}
