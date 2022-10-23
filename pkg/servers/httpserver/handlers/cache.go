package handlers

import (
	"context"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"net/http"
)

// Cache clear
// ?links=..... (,) - uid-s blocks and pages
// ?links=all - clear full cache
// @Router /cacheclear [get]
func (h *handlers) Cache(w http.ResponseWriter, r *http.Request) {
	in, err := cacheDecodeRequest(r.Context(), r)
	if err != nil {
		h.transportError(w, 500, err, "[Cache] Error function execution (CacheDecodeRequest)")
		return
	}
	serviceResult, err := h.service.Cache(r.Context(), in)
	if err != nil {
		h.transportError(w, 500, err, "[Cache] Error function execution (Cache)")
		return
	}
	response, _ := cacheEncodeResponse(r.Context(), serviceResult)
	if err != nil {
		h.transportError(w, 500, err, "[Cache] Error function execution (CacheEncodeResponse)")
		return
	}
	err = h.transportResponse(w, response)
	if err != nil {
		h.transportError(w, 500, err, "[Page] Error function execution (transportResponse)")
		return
	}

	return
}

func cacheDecodeRequest(ctx context.Context, r *http.Request) (in model.ServiceCacheIn, err error)  {
	in.Link = r.FormValue("link")
	if in.Link == "" {
		in.Link = r.FormValue("links")
	}

	return in, err
}

func cacheEncodeResponse(ctx context.Context, serviceResult model.RestStatus) (response interface{}, err error)  {
	return serviceResult, err
}