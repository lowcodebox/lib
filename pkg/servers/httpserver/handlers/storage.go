package handlers

import (
	"context"
	"net/http"
	"strings"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

func (h *handlers) Storage(w http.ResponseWriter, r *http.Request) {
	in, err := storageDecodeRequest(r.Context(), r)
	if err != nil {
		logger.Error(r.Context(), "[Alive] Error function execution (storageDecodeRequest).", zap.Error(err))
		return
	}
	serviceResult, err := h.service.Storage(r.Context(), in)
	if err != nil {
		logger.Error(r.Context(), "[Alive] Error function execution (service.Alive).", zap.Error(err))
		h.transportError(r.Context(), w, 500, err, "[Alive] Error service execution (service.Alive).")
		return
	}
	response, _ := storageEncodeResponse(r.Context(), serviceResult)
	if err != nil {
		logger.Error(r.Context(), "[Alive] Error function execution (storageEncodeResponse).", zap.Error(err))
		h.transportError(r.Context(), w, 500, err, "[Alive] Error function execution (storageEncodeResponse)")
		return
	}

	err = h.transportByte(w, response.MimeType, response.Body)
	if err != nil {
		h.transportError(r.Context(), w, 500, err, "[Query] Error function execution (transportResponse)")
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
