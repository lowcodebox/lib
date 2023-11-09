package handlers

import (
	"context"
	"net/http"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

// Cache clear
// ?links=..... (,) - uid-s blocks and pages
// ?links=all - clear full cache
// @Router /cacheclear [get]
func (h *handlers) Cache(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Alive] Error response execution", zap.Error(err))
		}
	}()

	in, er := cacheDecodeRequest(r.Context(), r)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Cache] error exec cacheDecodeRequest")
		return
	}

	serviceResult, err := h.service.Cache(r.Context(), in)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Cache] error exec service.Cache")
		return
	}

	response, _ := cacheEncodeResponse(r.Context(), serviceResult)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Cache] error exec cacheEncodeResponse")
		return
	}

	err = h.transportResponse(w, response)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Cache] error exec transportResponse")
		return
	}

	return
}

func cacheDecodeRequest(ctx context.Context, r *http.Request) (in model.ServiceCacheIn, err error) {
	in.Link = r.FormValue("link")
	if in.Link == "" {
		in.Link = r.FormValue("links")
	}

	return in, err
}

func cacheEncodeResponse(ctx context.Context, serviceResult model.RestStatus) (response interface{}, err error) {
	return serviceResult, err
}
