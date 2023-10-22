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
	_, err := pingDecodeRequest(r.Context(), r)
	if err != nil {
		logger.Error(r.Context(), "[Ping] Error function execution (PLoginDecodeRequest).", zap.Error(err))
		return
	}
	serviceResult, err := h.service.Ping(r.Context())
	if err != nil {
		logger.Error(r.Context(), "[Ping] Error function execution (service.Ping).", zap.Error(err))
		return
	}
	response, _ := pingEncodeResponse(r.Context(), serviceResult)
	if err != nil {
		logger.Error(r.Context(), "[Ping] Error function execution (PLoginEncodeResponse).", zap.Error(err))
		return
	}
	err = pingTransportResponse(w, response)
	if err != nil {
		logger.Error(r.Context(), "[Ping] Error function execution (PLoginTransportResponse).", zap.Error(err))
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

func pingTransportResponse(w http.ResponseWriter, response interface{}) (err error) {
	d, err := json.Marshal(response)

	w.Write(d)
	return err
}
