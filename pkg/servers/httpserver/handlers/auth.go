package handlers

import (
	"context"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"net/http"
)

// changerole get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /api/v1/changerole [get]
func (h *handlers) AuthChangeRole(w http.ResponseWriter, r *http.Request) {
	in, err := h.changeroleDecodeRequest(r.Context(), r)
	if err != nil {
		h.logger.Error(err, "[changerole] Error function execution (changeroleDecodeRequest).")
		return
	}
	serviceResult, err := h.service.AuthChangeRole(r.Context(), in)
	if err != nil {
		h.logger.Error(err, "[changerole] Error service execution (service.AuthChangeRole).")
		return
	}
	out, _ := h.changeroleEncodeResponse(r.Context(), serviceResult, in)
	if err != nil {
		h.logger.Error(err, "[changerole] Error function execution (changeroleEncodeResponse).")
		return
	}
	err = h.changeroleTransportResponse(w, r, out)
	if err != nil {
		h.logger.Error(err, "[changerole] Error function execution (changeroleTransportResponse).")
		return
	}

	return
}

func (h *handlers) changeroleDecodeRequest(ctx context.Context, r *http.Request) (request model.ServiceAuthIn, err error)  {
	request.Profile = r.FormValue("profile")	// uid-профиля, который надо сделать активным
	expire	:= r.FormValue("expire")	// признак того, что надо вернуть протухших, но валидный токен

	if expire == "true" || expire == "1" {
		request.Expire = true
	}

	return request, err
}

func (h *handlers) changeroleEncodeResponse(ctx context.Context, serviceResult model.ServiceAuthOut, in model.ServiceAuthIn) (response model.ServiceAuthOut, err error)  {
	response.RequestURI = in.RequestURI
	response.Token = serviceResult.Token
	return serviceResult, err
}

func (h *handlers) changeroleTransportResponse(w http.ResponseWriter, r *http.Request, out model.ServiceAuthOut) (err error)  {
	// редиректим страницу, передав в куку новый токен с просроченным временем
	w.Header().Set("X-Auth-Key", out.Token)

	cookie := &http.Cookie{
		Path: "/",
		Name:   "X-Auth-Key",
		Value:  out.Token,
		MaxAge: 30000,
	}

	//// переписываем куку у клиента
	http.SetCookie(w, cookie)
	http.Redirect(w, r, r.Referer(), 302)

	return err
}
