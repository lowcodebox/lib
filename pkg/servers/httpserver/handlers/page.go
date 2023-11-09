package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Page get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /api/v1/page [get]
func (h *handlers) Page(w http.ResponseWriter, r *http.Request) {
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Alive] Error response execution", zap.Error(err))
		}
	}()

	in, er := pageDecodeRequest(r.Context(), r)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Page] error exec pageDecodeRequest")
		return
	}

	serviceResult, err := h.service.Page(r.Context(), in)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Page] error exec service.Page")
		return
	}

	response, er := pageEncodeResponse(r.Context(), &serviceResult)
	if er != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Page] error exec pageEncodeResponse")
		return
	}

	err = h.transportResponseHTTP(w, response)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Page] error exec transportResponseHTTP")
		return
	}

	return
}

func pageDecodeRequest(ctx context.Context, r *http.Request) (in model.ServiceIn, err error) {
	vars := mux.Vars(r)
	in.Page = vars["page"]
	err = r.ParseForm()
	if err != nil {
		logger.Error(ctx, "[Block] (blockDecodeRequest) error ParseForm", zap.Error(err))
	}

	in.Url = r.URL.Query().Encode()
	in.Referer = r.Referer()
	in.RequestURI = r.RequestURI
	in.QueryRaw = r.URL.RawQuery
	in.Form = r.Form
	in.PostForm = r.PostForm
	in.Host = r.Host
	in.Method = r.Method
	in.Query = r.URL.Query()
	in.RequestRaw = r

	slURI := strings.Split(in.RequestURI, "?")
	in.CachePath = slURI[0]
	if len(slURI) > 1 {
		in.CacheQuery = slURI[1]
	}

	// указатель на профиль текущего пользователя
	var profile models.ProfileData
	profileRaw := r.Context().Value("UserRaw")
	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)

	profile, ok := ctx.Value("profile").(models.ProfileData)
	if ok {
		in.Profile = profile
	}
	in.Profile = profile

	return in, err
}

func pageEncodeResponse(ctx context.Context, serviceResult *model.ServicePageOut) (response string, err error) {
	response = serviceResult.Body
	return response, err
}
