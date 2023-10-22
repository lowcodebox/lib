package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/models"
	"github.com/gorilla/mux"
)

// Block get user by login+pass pair
// @Summary get user by login+pass pair
// @Param login_input body model.Pong true "login data"
// @Success 200 {object} model.Pong [Result:model.Pong]
// @Failure 400 {object} model.Pong
// @Failure 500 {object} model.Pong
// @Router /api/v1/block [get]
func (h *handlers) Block(w http.ResponseWriter, r *http.Request) {
	in, err := blockDecodeRequest(r.Context(), r)
	if err != nil {
		h.transportError(r.Context(), w, 500, err, "[Block] Error function execution (BlockDecodeRequest)")
		return
	}
	serviceResult, err := h.service.Block(r.Context(), in)
	if err != nil {
		h.transportError(r.Context(), w, 500, err, "[Block] Error function execution (Block)")
		return
	}
	response, _ := blockEncodeResponse(r.Context(), &serviceResult)
	if err != nil {
		h.transportError(r.Context(), w, 500, err, "[Block] Error function execution (BlockEncodeResponse)")
		return
	}
	err = h.transportResponseHTTP(w, string(response))
	if err != nil {
		h.transportError(r.Context(), w, 500, err, "[Page] Error function execution (transportResponse)")
		return
	}

	return
}

func blockDecodeRequest(ctx context.Context, r *http.Request) (in model.ServiceIn, err error) {
	vars := mux.Vars(r)
	in.Block = vars["block"]
	r.ParseForm()

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

	profile, ok := ctx.Value("profile").(models.ProfileData)
	if ok {
		in.Profile = profile
	}

	fmt.Println(profile, ok, ctx.Value("profile"))

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
