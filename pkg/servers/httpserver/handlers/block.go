package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"git.edtech.vm.prod-6.cloud.el/fabric/app/pkg/model"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Block get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /api/v1/block [get]
func (h *handlers) Block(w http.ResponseWriter, r *http.Request) {
	var serviceResult model.ServiceBlockOut
	var in model.ServiceIn
	var response template.HTML
	var err error
	defer func() {
		if err != nil {
			logger.Error(h.ctx, "[Block] Error response execution",
				zap.String("url", r.RequestURI),
				zap.String("in", fmt.Sprintf("%+v", in)),
				zap.Error(err))
		}
	}()

	in, err = blockDecodeRequest(r.Context(), r)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Block] error exec blockDecodeRequest")
		return
	}

	serviceResult, err = lib.Retrier(h.cfg.MaxCountRetries.Value, h.cfg.TimeRetries.Value, true, func() (model.ServiceBlockOut, error) {
		serviceResult, err = h.service.Block(r.Context(), in)
		return serviceResult, err
	})
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Block] error exec service.Block")
		return
	}

	response, err = blockEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Block] error exec blockEncodeResponse")
		return
	}

	err = h.transportResponseHTTP(w, string(response))
	if err != nil {
		err = h.transportError(r.Context(), w, 500, err, "[Block] error exec transportResponseHTTP")
		return
	}

	return
}

func blockDecodeRequest(ctx context.Context, r *http.Request) (in model.ServiceIn, err error) {
	vars := mux.Vars(r)
	in.Block = vars["block"]
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
	in.CacheSkip = r.FormValue("skip_cache") // true/false

	in.RequestRaw = r

	profile, ok := ctx.Value("profile").(models.ProfileData)
	if ok {
		in.Profile = profile
	}

	//fmt.Println(profile, ok, ctx.Value("profile"))

	//cookieCurrent, err := r.Cookie("sessionID")
	//iam := ""
	//if err == nil {
	//	tokenI := strings.Split(fmt.Sprint(cookieCurrent), "=")
	//	if len(tokenI) > 1 {
	//		iam = tokenI[1]
	//	}
	//}
	//in.Token = iam
	//
	//// указатель на профиль текущего пользователя
	//var profile model.ProfileData
	//profileRaw := r.Context().Value("UserRaw")
	//json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)
	//
	//in.Profile = profile

	return in, err
}

func blockEncodeResponse(ctx context.Context, serviceResult *model.ServiceBlockOut) (response template.HTML, err error) {
	response = serviceResult.Result
	return response, err
}
