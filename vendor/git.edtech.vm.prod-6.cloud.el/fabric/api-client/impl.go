package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"git.edtech.vm.prod-6.cloud.el/packages/curl"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"go.uber.org/zap"
)

func (a *api) data(ctx context.Context, tpls, option, role, page, size string) (result models.ResponseData, err error) {
	if tpls == "" {
		return result, fmt.Errorf("error request to orm (Data). err: param tpls in empty")
	}

	var headers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	headers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "Data", err)
	}

	urlc, err := url.JoinPath(a.url, "data", tpls)
	if err != nil {
		return result, fmt.Errorf("error request to orm (Data). err: %w)", err)
	}

	if option != "" {
		urlc, err = url.JoinPath(urlc, option)
		if err != nil {
			return result, fmt.Errorf("error request to orm (Data). err: %w)", err)
		}
	}

	urlObj, err := url.Parse(urlc)
	if err != nil {
		return result, fmt.Errorf("error request to orm (Data). err: %w)", err)
	}

	// Добавление query параметров
	query := urlObj.Query()
	if role != "" {
		query.Add("role", role)
	}
	if page != "" {
		query.Add("page", page)
	}
	if size != "" {
		query.Add("size", size)
	}
	urlObj.RawQuery = query.Encode()

	_, err = lib.Curl(ctx, http.MethodPost, urlObj.String(), "", &result, headers, nil)
	if err != nil {
		err = fmt.Errorf("error request to orm (Data). err: %w, urlc: %s, method: %s", err, urlc, http.MethodPost)
	}

	return result, err
}

// Query результат выводим в объект как при вызове Curl
func (a *api) query(ctx context.Context, query, method, bodyJSON, group string) (result string, err error) {
	var handlers = map[string]string{}
	var cookies []*http.Cookie
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token

	if group != "" {
		cookies = append(cookies, &http.Cookie{
			Path:     "/",
			Name:     "groupID",
			Value:    group,
			MaxAge:   30000,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
	}

	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "Query", err, query, method, bodyJSON)
	}

	urlc := a.url + "/query/" + query
	urlc = strings.Replace(urlc, "//query", "/query", 1)

	res, err := lib.Curl(ctx, method, urlc, bodyJSON, nil, handlers, cookies)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return fmt.Sprint(res), err
}

func (a *api) tpls(ctx context.Context, role, option string) (result models.ResponseData, err error) {
	if role == "" {
		role = "_all"
	}

	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "tpls", err, role, option)
	}

	urlc := a.url
	if option == "" {
		urlc, err = url.JoinPath(urlc, "tpls", role)
	} else {
		urlc, err = url.JoinPath(urlc, "tpls", role, option)
	}
	if err != nil {
		return result, fmt.Errorf("error request to api (Tpls). err: %w)", err)
	}

	_, err = lib.Curl(ctx, "GET", urlc, "", &result, handlers, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

// search результат выводим в объект как при вызове Curl
func (a *api) search(ctx context.Context, query, method, bodyJSON string) (resp string, err error) {
	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return resp, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token

	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "Search", err, query, method, bodyJSON)
	}

	urlc := a.url + "/search"
	urlc = strings.Replace(urlc, "//search", "/search", 1)

	res, err := lib.Curl(ctx, method, urlc, bodyJSON, nil, handlers, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return fmt.Sprint(res), err
}

func (a *api) objGet(ctx context.Context, uids string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "ObjGet", err, uids)
	}

	urlc := a.url + "/objs/" + uids
	urlc = strings.Replace(urlc, a.url+"//objs/", a.url+"/objs/", 1)

	_, err = lib.Curl(ctx, "GET", urlc, "", &result, handlers, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

func (a *api) linkGet(ctx context.Context, tpl, obj, mode, short string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "LinkGet", err, tpl, obj, mode, short)
	}

	urlc := a.url + "/link/get?source=" + tpl + "&mode=" + mode + "&obj=" + obj + "&short=" + short
	urlc = strings.Replace(urlc, "//link", "/link", 1)

	_, err = lib.Curl(ctx, "GET", urlc, "", &result, handlers, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

func (a *api) linkOperation(ctx context.Context, operation, element, from, to string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "linkAdd", err, element, from, to)
	}

	urlc := a.url + "/link/" + operation + "?element=" + element + "&from=" + from + "&to=" + to
	urlc = strings.Replace(urlc, "//link", "/link", 1)

	if operation != "add" && operation != "delete" {
		err = fmt.Errorf("operation '%s' is not resolved. (url: %s)", operation, urlc)
		return result, err
	}

	_, err = lib.Curl(ctx, "GET", urlc, "", &result, handlers, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

// ObjAttrUpdate изменение значения аттрибута объекта
func (a *api) objAttrUpdate(ctx context.Context, uid, name, value, src, editor string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "ObjAttrUpdate", err, uid, name, value, src, editor)
	}

	post := map[string]string{}
	thisTime := fmt.Sprintf("%v", time.Now().UTC())

	post["uid"] = uid
	post["element"] = name
	post["value"] = value
	post["src"] = src
	post["rev"] = thisTime
	post["path"] = ""
	post["token"] = ""
	post["editor"] = editor

	dataJ, _ := json.Marshal(post)
	result, err = a.Element(ctx, "update", string(dataJ))

	return result, err
}

// Element
// TODO ПЕРЕДЕЛАТЬ на понятные пути в ORM
// сделано так для совместимости со старой версией GUI
// Action:
// block - Блокировка поля для его редактировании
// unblock - Разблокировка заблокированного поля
// update - Обновление значения заблокированного поля
// checkup - Проверяем переданное значение на соответствие выбранному условию
// all (elements) - Получаем поля, по заданному в параметрах типу
// "" - без действия - получаем все поля для объекта
func (a *api) element(ctx context.Context, action, body string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	var urlc string
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "Element", err, urlc, action, body)
	}

	// получаем поля шаблона
	if action == "elements" || action == "all" {
		urlc = a.url + "/element/" + body
		urlc = strings.Replace(urlc, "//element", "/element", 1)

		_, err = lib.Curl(ctx, "GET", urlc, "", &result, handlers, nil)
		if err != nil {
			err = fmt.Errorf("%s (url: %s)", err, a.url+"/element/"+body)
		}
		return result, err
	}

	urlc = a.url + "/element/" + action + "?format=json"
	urlc = strings.Replace(urlc, "//element", "/element", 1)

	_, err = curl.NewRequestDefault().Method("POST").Payload(body).MapToObj(&result).Headers(handlers).Url(urlc).Do(nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

func (a *api) objCreate(ctx context.Context, bodymap map[string]string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	var urlc string
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "ObjCreate", err, urlc, bodymap)
	}

	body, err := json.Marshal(bodymap)
	if err != nil {
		return result, fmt.Errorf("error Marshal. bodymap: %+v, err: %v", bodymap, err)
	}

	urlc = a.url + "/objs?format=json"
	urlc = strings.Replace(urlc, "//objs", "/objs", 1)

	res, err := curl.NewRequestDefault().Method("POST").Payload(string(body)).MapToObj(&result).Headers(handlers).Url(urlc).Do(nil)
	if err != nil {
		result.Res = res
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	//res, err := lib.Curl(ctx, "POST", urlc, string(body), &result, handlers, nil)
	//if err != nil {
	//	result.Res = res
	//	err = fmt.Errorf("%s (url: %s)", err, urlc)
	//}

	return result, err
}

func (a *api) objDelete(ctx context.Context, uids string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	var urlc string
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "objDelete", err, urlc, uids)
	}

	payload := map[string]string{
		"data-uid": uids,
	}
	payloadJson, err := json.Marshal(payload)

	urlc = a.url + "/objs/delete?ids=" + uids
	urlc = strings.Replace(urlc, "//objs", "/objs", 1)

	_, err = lib.Curl(ctx, "JSONTOPOST", urlc, string(payloadJson), &result, handlers, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

func (a *api) tools(ctx context.Context, method, action string, params map[string]interface{}) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	if err != nil {
		return result, fmt.Errorf("error GenXServiceKey. err: %s", err)
	}
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "Tools", err, method, action, params)
	}

	urlc, err := url.JoinPath(a.url, "tools", action)
	if err != nil {
		return result, err
	}
	bodyJSON := ""
	// Если json - все параметры как тело, иначе как query
	if method == http.MethodPost {
		bytes, err := json.Marshal(params)
		if err != nil {
			return result, err
		}
		bodyJSON = string(bytes)
	} else {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, fmt.Sprint(v))
		}
		urlc = urlc + "?" + q.Encode()
	}

	_, err = lib.Curl(ctx, method, urlc, bodyJSON, &result, handlers, nil)
	if err != nil {
		err = fmt.Errorf("%w (url: %s)", err, urlc)
	}

	return result, err
}

func (a *api) observeLogger(ctx context.Context, start time.Time, method string, err error, arguments ...interface{}) {
	logger.Info(ctx, "timing api query",
		zap.String("method", method),
		zap.Float64("timing", time.Since(start).Seconds()),
		//zap.String("arguments", fmt.Sprint(arguments)),
		zap.Error(err),
	)
}
