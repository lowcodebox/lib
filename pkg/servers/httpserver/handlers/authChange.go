package handlers

import (
	"context"
	"net/http"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"go.uber.org/zap"
)

func (h *handlers) AuthChangeRole(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[AuthChangeRole] Error response execution",
				zap.String("url", r.RequestURI),
				zap.Error(err))
		}
	}()

	in, er := h.changeroleDecodeRequest(r.Context(), r)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[AuthChangeRole] error exec changeroleDecodeRequest")
		return
	}

	serviceResult, er := h.service.AuthChangeRole(r.Context(), in)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[AuthChangeRole] error exec AuthChangeRole")
		return
	}

	out, er := h.changeroleEncodeResponse(r.Context(), serviceResult, in)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[AuthChangeRole] error exec changeroleEncodeResponse")
		return
	}

	err = h.changeroleTransportResponse(w, r, out)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, er, "[AuthChangeRole] error exec changeroleTransportResponse")
		return
	}

	return
}

func (h *handlers) changeroleDecodeRequest(ctx context.Context, r *http.Request) (request model.ServiceAuthChangeIn, err error) {
	request.Profile = r.FormValue("profile") // uid-профиля, который надо сделать активным
	expire := r.FormValue("expire")          // признак того, что надо вернуть протухших, но валидный токен

	if expire == "true" || expire == "1" {
		request.Expire = true
	}

	return request, err
}

func (h *handlers) changeroleEncodeResponse(ctx context.Context, serviceResult model.ServiceAuthChangeOut, in model.ServiceAuthChangeIn) (response model.ServiceAuthChangeOut, err error) {
	response.RequestURI = in.RequestURI
	response.Token = serviceResult.Token
	return serviceResult, err
}

func (h *handlers) changeroleTransportResponse(w http.ResponseWriter, r *http.Request, out model.ServiceAuthChangeOut) (err error) {
	// редиректим страницу, передав в куку новый токен с просроченным временем
	w.Header().Set("X-Auth-Key", out.Token)

	cookie := &http.Cookie{
		Path:     "/",
		Name:     "X-Auth-Key",
		Value:    out.Token,
		MaxAge:   5256000,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	//// переписываем куку у клиента
	http.SetCookie(w, cookie)
	http.Redirect(w, r, r.Referer(), 302)

	return err
}
