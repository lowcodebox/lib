package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"go.uber.org/zap"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
)

const authTokenName = "X-Auth-Key"

type authResponse struct {
	XAuthToken  string `json:"x_auth_token"`
	UserUID     string `json:"user_uid"`
	ProfileUID  string `json:"profile_uid"`
	Ref         string `json:"ref"`
	Code        string `json:"code"`
	Description string `json:"description"`
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

	// устанавливаем сервисные куки после авторизации (для фронта)
	for _, v := range strings.Split(h.cfg.CookieFrontLogin, ",") {
		if v == "" {
			continue
		}

		nv := strings.Split(v, "=")
		if len(nv) < 2 {
			continue
		}
		name := nv[0]
		value := nv[1]

		// заменяем куку у пользователя в браузере
		c := &http.Cookie{
			Path:     "/",
			Name:     name,
			Value:    value,
			MaxAge:   5256000,
			HttpOnly: false,
			Secure:   false,
		}

		http.SetCookie(w, c)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(r.Context(), "error. readAll request is failed", zap.Error(err))
		return in, fmt.Errorf("error. readAll request is failed. err: %s", err)
	}

	defer r.Body.Close()

	in.Payload = string(body)

	return in, err
}

func (h *handlers) authEncodeResponse(ctx context.Context, serviceResult model.ServiceAuthOut) (response authResponse, err error) {
	response.XAuthToken = serviceResult.XAuthToken
	response.UserUID = serviceResult.UserUID
	response.ProfileUID = serviceResult.ProfileUID
	response.Ref = serviceResult.Ref

	if serviceResult.Error == nil {
		response.Code = "Success"
		response.Description = "Авторизация успешна"
	}
	return response, err
}

func (h *handlers) authTransportResponse(w http.ResponseWriter, r *http.Request, out authResponse) (err error) {
	token := fmt.Sprint(out.XAuthToken)

	// редиректим страницу, передав в куку новый токен с просроченным временем
	w.Header().Set(authTokenName, token)

	cookie := &http.Cookie{
		Path:     "/",
		Name:     authTokenName,
		Value:    token,
		MaxAge:   5256000,
		HttpOnly: true,
		Secure:   false,
		SameSite: h.cfg.GetCookieSameSite(),
	}

	//// переписываем куку у клиента
	http.SetCookie(w, cookie)
	//http.Redirect(w, r, r.Referer(), 302)

	return err
}

func (h *handlers) deleteCookie(w http.ResponseWriter, r *http.Request) (err error) {
	w.Header().Set("X-Auth-Key", "")

	setExpiredCookie(w, authTokenName)
	setExpiredCookie(w, h.cfg.NameCookieWBTokenV3)
	setExpiredCookie(w, h.cfg.NameCookieWbxValidationKey)

	// удаляем сервисные куки если не авторизован (для фронта)
	for _, name := range strings.Split(h.cfg.CookieFrontLogoutDelete, ",") {
		if name == "" {
			continue
		}
		setExpiredCookie(w, name)
	}

	http.Redirect(w, r, r.Referer(), http.StatusFound)

	return err
}

func setExpiredCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:    name,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0), // Устанавливаем истечение
	}

	//// переписываем куку у клиента
	http.SetCookie(w, cookie)
}
