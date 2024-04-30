package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"github.com/golang-jwt/jwt"
)

// Refresh отправляем старый X-Auth-Access-токен
// получаем X-Auth-Access токен (два токена + текущая авторизационная сессия)
// этот ключ добавляется в куки или сохраняется в сервисе как ключ доступа
// profile - uid-профиля, под которым проводим авторизацию
// expire - признак того, что refresh-токен прийдет просроченный в новом токене
func (o *iam) refresh(ctx context.Context, token, profile string, expire bool) (result string, err error) {
	var res models.Response
	var handlers = map[string]string{}
	serviceKey, err := lib.GenXServiceKey(o.domain, []byte(o.projectKey), tokenInterval)
	handlers[headerServiceKey] = serviceKey
	if o.observeLog {
		defer o.observeLogger(ctx, time.Now(), "refresh", err, token, profile, expire)
	}

	urlc := o.url + "/token/refresh?profile=" + profile + "&expire=" + fmt.Sprint(expire)
	urlc = strings.Replace(urlc, "//token", "/token", 1)

	_, err = lib.Curl(ctx, "POST", urlc, token, &res, map[string]string{}, nil)
	if err != nil {
		return result, fmt.Errorf("urlc: %s, err: %s", urlc, err)
	}

	result = fmt.Sprint(res.Data)

	return result, err
}

func (o *iam) profileGet(ctx context.Context, sessionID string) (result string, err error) {
	var res models.Response
	var handlers = map[string]string{}
	serviceKey, err := lib.GenXServiceKey(o.domain, []byte(o.projectKey), tokenInterval)
	handlers[headerServiceKey] = serviceKey
	if o.observeLog {
		defer o.observeLogger(ctx, time.Now(), "refresh", err, sessionID)
	}

	urlc := o.url + "/profile/" + sessionID
	urlc = strings.Replace(urlc, "//profile", "/profile", 1)

	_, err = lib.Curl(ctx, "GET", urlc, "", &res, handlers, nil)
	if err != nil {
		return result, fmt.Errorf("urlc: %s, err: %s", urlc, err)
	}

	b2, _ := json.Marshal(res.Data)

	return string(b2), err
}

func (o *iam) profileList(ctx context.Context) (result string, err error) {
	var res models.Response
	var handlers = map[string]string{}
	serviceKey, err := lib.GenXServiceKey(o.domain, []byte(o.projectKey), tokenInterval)
	handlers[headerServiceKey] = serviceKey
	if o.observeLog {
		defer o.observeLogger(ctx, time.Now(), "profileList", err)
	}

	urlc := o.url + "/profile/list"
	urlc = strings.Replace(urlc, "//profile", "/profile", 1)

	_, err = lib.Curl(ctx, "GET", urlc, "", &res, handlers, nil)
	if err != nil {
		return result, fmt.Errorf("urlc: %s, err: %s", urlc, err)
	}

	result = fmt.Sprint(res.Data)

	return result, err
}

func (o *iam) auth(ctx context.Context, suser, ref string) (status bool, token string, err error) {
	var res models.Response
	var handlers = map[string]string{}
	serviceKey, err := lib.GenXServiceKey(o.domain, []byte(o.projectKey), tokenInterval)
	handlers[headerServiceKey] = serviceKey
	if o.observeLog {
		defer func() {
			o.observeLogger(ctx, time.Now(), "auth", err, token)
		}()
	}

	urlc := o.url + "/auth?suser=&ref=" + ref
	urlc = strings.Replace(urlc, "//auth", "/auth", 1)

	_, err = lib.Curl(ctx, http.MethodPost, urlc, suser, &res, handlers, nil)
	if err != nil {
		return false, "", fmt.Errorf("urlc: %s, err: %s", urlc, err)
	}

	return true, fmt.Sprint(res.Data), nil
}

func (o *iam) verify(ctx context.Context, tokenString string) (status bool, body *models.Token, refreshToken string, err error) {
	var in models.Token
	var jtoken = map[string]string{}

	jsonToken, err := lib.Decrypt([]byte(o.projectKey), tokenString)
	if err != nil {
		return false, nil, refreshToken, err
	}

	err = json.Unmarshal([]byte(jsonToken), &jtoken)
	if err != nil {
		return false, nil, refreshToken, err
	}

	tokenAccess := jtoken["access"]
	refreshToken = jtoken["refresh"]

	token, err := jwt.ParseWithClaims(tokenAccess, &in, func(token *jwt.Token) (interface{}, error) {
		return []byte(o.projectKey), nil
	})

	if !token.Valid {
		return false, nil, refreshToken, o.msg.TokenValidateFail.Error("")
	}
	tbody := token.Claims.(*models.Token)

	return true, tbody, refreshToken, err
}
