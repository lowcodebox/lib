package handlers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

const authTokenName = "X-Auth-Key"

type authResponse struct {
	models.Response
	Ref string `json:"ref"`
}

func (h *handlers) AuthLogOut(w http.ResponseWriter, r *http.Request) {
	var err error

	err = h.deleteCookie(w, r)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[AuthLogOut] error delete cookie")
		return
	}

	return
}

func (h *handlers) AuthLogIn(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[AuthLogIn] Error response execution", zap.Error(err))
		}
	}()

	ctx := r.Context()
	err = h.localization(w, r)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[AuthLogIn] error exec localization")
		return
	}

	in, er := h.authDecodeRequest(&ctx, r)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[AuthLog] error exec authDecodeRequest")
		return
	}

	serviceResult, er := h.service.AuthLogIn(r.Context(), in)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[AuthLog] error exec AuthLog")
		return
	}

	out, er := h.authEncodeResponse(r.Context(), serviceResult)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, er, "[AuthLog] error exec authEncodeResponse")
		return
	}

	err = h.authTransportResponse(w, r, out)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, er, "[AuthLog] error exec authTransportResponse")
		return
	}

	err = h.transportResponse(w, out)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[AuthLog] error exec transportResponse")
		return
	}

	return
}

// и при регистрации, и при авторизации через внешних сервис мы отправляем в едином формате строку в гете
func (h *handlers) authDecodeRequest(ctx *context.Context, r *http.Request) (in model.ServiceAuthIn, err error) {
	var localization = "EN"

	// проверяем куку на наличие локализации
	cookie, err := r.Cookie("local")
	if err == nil {
		localization = cookie.Value
	}

	*ctx = context.WithValue(*ctx, "Local", localization)

	// параметры могут быть в GET-запросе (для совместимости со старыми версиями), но по-новому в теле POST-а
	suser, err := url.QueryUnescape(r.FormValue("suser"))
	if err != nil {
		return in, fmt.Errorf("error. QueryUnescape, string: %s, err: %s", r.FormValue("suser"), err)
	}

	in.Ref = r.FormValue("ref")
	if suser != "" {
		in.Payload = suser
		return in, nil
	}

	// если передали в json в теле post-а
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Error(r.Context(), "error. readAll request is failed", zap.Error(err))
		return in, fmt.Errorf("error. readAll request is failed. err: %s", err)
	}

	defer r.Body.Close()

	in.Payload = string(body)

	return in, err
}

func (h *handlers) authEncodeResponse(ctx context.Context, serviceResult model.ServiceAuthOut) (response authResponse, err error) {
	response.Data = serviceResult.XAuthToken
	response.Ref = serviceResult.Ref

	if serviceResult.Error == nil {
		response.Status.Code = "Success"
		response.Status.Description = "Авторизация успешна"
	}
	return response, err
}

func (h *handlers) authTransportResponse(w http.ResponseWriter, r *http.Request, out authResponse) (err error) {
	token := fmt.Sprint(out.Response.Data)

	// редиректим страницу, передав в куку новый токен с просроченным временем
	w.Header().Set(authTokenName, token)

	cookie := &http.Cookie{
		Path:     "/",
		Name:     authTokenName,
		Value:    token,
		MaxAge:   30000,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	//// переписываем куку у клиента
	http.SetCookie(w, cookie)
	//http.Redirect(w, r, r.Referer(), 302)

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
