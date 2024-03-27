package handlers

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

func (h *handlers) Storage(w http.ResponseWriter, r *http.Request) {
	var err error
	var in model.StorageIn
	var serviceResult, response model.StorageOut

	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Storage] Error response execution",
				zap.String("url", r.RequestURI),
				zap.String("in", fmt.Sprintf("%+v", in)),
				zap.Error(err))
		}
	}()

	in, err = storageDecodeRequest(r.Context(), r)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Storage] error exec storageDecodeRequest")
		return
	}

	serviceResult, err = h.service.Storage(r.Context(), in)
	if err != nil {
		err = h.transportError(r.Context(), w, 404, err, "[Storage] error exec service.Storage")
		return
	}

	response, err = storageEncodeResponse(r.Context(), serviceResult)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Storage] error exec storageEncodeResponse")
		return
	}

	err = h.transportByte(w, response.MimeType, response.Body)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Storage] error exec transportByte")
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

	file = filepath.Clean(file)
	//if !strings.Contains(file, "/assets/") && !strings.Contains(file, "/templates/") {
	//	return request, fmt.Errorf("error. path is not valid. file: %s", file)
	//}

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
