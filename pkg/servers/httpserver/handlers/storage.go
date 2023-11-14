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
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Alive] Error response execution", zap.Error(err))
		}
	}()

	in, er := storageDecodeRequest(r.Context(), r)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Storage] error exec storageDecodeRequest")
		return
	}

	serviceResult, err := h.service.Storage(r.Context(), in)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Storage] error exec service.Storage")
		return
	}

	response, _ := storageEncodeResponse(r.Context(), serviceResult)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Storage] error exec storageEncodeResponse")
		return
	}

	err = h.transportByte(w, response.MimeType, response.Body)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Storage] error exec transportByte")
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
