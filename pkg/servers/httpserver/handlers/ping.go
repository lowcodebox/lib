package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

// Ping get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /api/v1/ping [get]
func (h *handlers) Ping(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Ping] Error response execution", zap.Error(err))
		}
	}()

	_, er := pingDecodeRequest(r.Context(), r)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Ping] error exec pingDecodeRequest")
		return
	}

	serviceResult, er := h.service.Ping(r.Context())
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Ping] error exec service.Ping")
		return
	}

	response, er := pingEncodeResponse(r.Context(), serviceResult)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Ping] error exec pingEncodeResponse")
		return
	}

	err = pingTransportResponse(r.Context(), w, response)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, er, "[Ping] error exec pingTransportResponse")
		return
	}

	return
}

func pingDecodeRequest(ctx context.Context, r *http.Request) (request *[]models.Pong, err error) {
	return request, err
}

func pingEncodeResponse(ctx context.Context, serviceResult []models.Pong) (response []models.Pong, err error) {
	return serviceResult, err
}

func pingTransportResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	d, err := json.Marshal(response)
	if err != nil {
		logger.Error(ctx, "[Ping] (pingTransportResponse) error ParseForm", zap.Error(err))
	}

	_, err = w.Write(d)
	if err != nil {
		logger.Error(ctx, "[Ping] (pingTransportResponse) error Write", zap.Error(err))
	}

	return err
}
