package handlers

import (
	"context"
	"net/http"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

// Alive get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /api/v1/alive [get]
func (h *handlers) Alive(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Alive] Error response execution",
				zap.String("url", r.RequestURI),
				zap.Error(err))
		}
	}()

	_, err = aliveDecodeRequest(r.Context(), r)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Alive] error exec aliveDecodeRequest")
		return
	}

	serviceResult, err := h.service.Alive(r.Context())
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Alive] error exec service.Alive")
		return
	}

	response, _ := aliveEncodeResponse(r.Context(), serviceResult)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Alive] error exec aliveEncodeResponse")
		return
	}

	err = h.transportResponse(w, response)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Alive] error exec transportResponse")
		return
	}
	return
}

func aliveDecodeRequest(ctx context.Context, r *http.Request) (request model.AliveOut, err error) {
	return request, err
}

func aliveEncodeResponse(ctx context.Context, serviceResult model.AliveOut) (response model.AliveOut, err error) {
	return serviceResult, err
}
