package handlers

import (
	"context"
	"net/http"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

func (h *handlers) AuthLogOut(w http.ResponseWriter, r *http.Request) {
	var err error

	err = h.deleteCookie(w, r)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[AuthLogOut] error delete cookie")
		return
	}

	return
}

func (h *handlers) AuthChangeRole(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Alive] Error response execution", zap.Error(err))
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

func (h *handlers) changeroleDecodeRequest(ctx context.Context, r *http.Request) (request model.ServiceAuthIn, err error) {
	request.Profile = r.FormValue("profile") // uid-профиля, который надо сделать активным
	expire := r.FormValue("expire")          // признак того, что надо вернуть протухших, но валидный токен

	if expire == "true" || expire == "1" {
		request.Expire = true
	}

	return request, err
}

func (h *handlers) changeroleEncodeResponse(ctx context.Context, serviceResult model.ServiceAuthOut, in model.ServiceAuthIn) (response model.ServiceAuthOut, err error) {
	response.RequestURI = in.RequestURI
	response.Token = serviceResult.Token
	return serviceResult, err
}

func (h *handlers) changeroleTransportResponse(w http.ResponseWriter, r *http.Request, out model.ServiceAuthOut) (err error) {
	// редиректим страницу, передав в куку новый токен с просроченным временем
	w.Header().Set("X-Auth-Key", out.Token)

	cookie := &http.Cookie{
		Path:     "/",
		Name:     "X-Auth-Key",
		Value:    out.Token,
		MaxAge:   30000,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	//// переписываем куку у клиента
	http.SetCookie(w, cookie)
	http.Redirect(w, r, r.Referer(), 302)

	return err
}

func (h *handlers) deleteCookie(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("X-Auth-Key", "")

	cookie := &http.Cookie{
		Path:    "/",
		Name:    "X-Auth-Key",
		Expires: time.Unix(0, 0),
		Value:   "",
		MaxAge:  30000,
		Secure:  true,
	}

	//// переписываем куку у клиента
	http.SetCookie(w, cookie)
	http.Redirect(w, r, r.Referer(), 302)

	return err
}
