package handlers

import (
	"context"
	"fmt"
	"net/http"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"go.uber.org/zap"
)

// Cache clear
// ?links=..... (,) - uid-s blocks and pages
// ?links=all - clear full cache
// @Router /cacheclear [get]
func (h *handlers) Cache(w http.ResponseWriter, r *http.Request) {
	var err error
	var in model.ServiceCacheIn
	var response interface{}
	var serviceResult model.RestStatus
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Cache] Error response execution",
				zap.String("in", fmt.Sprintf("%+v", in)),
				zap.String("url", r.RequestURI),
				zap.Error(err))
		}
	}()

	in, err = cacheDecodeRequest(r.Context(), r)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Cache] error exec cacheDecodeRequest")
		return
	}

	serviceResult, err = h.service.Cache(r.Context(), in)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Cache] error exec service.Cache")
		return
	}

	response, err = cacheEncodeResponse(r.Context(), serviceResult)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Cache] error exec cacheEncodeResponse")
		return
	}

	err = h.transportResponse(w, response)
	if err != nil {
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
