package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

// Files operation from files
// @Router /api/v1/files [get/post/put/delete]
func (h *handlers) Files(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Files] Error response execution",
				zap.String("url", r.RequestURI),
				zap.Error(err))
		}
	}()

	in, er := filesDecodeRequest(r.Context(), r)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Files] error exec filesDecodeRequest")
		return
	}

	serviceResult, er := h.service.Files(r.Context(), in)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Files] error exec service.Files")
		return
	}

	response, er := filesEncodeResponse(r.Context(), serviceResult)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Files] error exec filesEncodeResponse")
		return
	}

	err = filesTransportResponse(r.Context(), w, response)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Files] error exec filesTransportResponse")
		return
	}

	return
}

func filesDecodeRequest(ctx context.Context, r *http.Request) (in model.ServiceFilesIn, err error) {
	if r.Method == http.MethodPost {
		in.Action = model.FilesActionLoad
	}
	return in, err
}

func filesEncodeResponse(ctx context.Context, serviceResult model.ServiceFilesOut) (response model.ServiceFilesOut, err error) {
	return serviceResult, err
}

func filesTransportResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	d, err := json.Marshal(response)
	if err != nil {
		logger.Error(ctx, "[Files] (filesTransportResponse) error ParseForm", zap.Error(err))
	}

	_, err = w.Write(d)
	if err != nil {
		logger.Error(ctx, "[Files] (filesTransportResponse) error Write", zap.Error(err))
	}

	return err
}
