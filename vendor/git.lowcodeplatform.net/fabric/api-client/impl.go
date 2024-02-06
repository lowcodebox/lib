package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/curl"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

// Query результат выводим в объект как при вызове Curl
func (a *api) query(ctx context.Context, query, method, bodyJSON string) (result string, err error) {
	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
	handlers[headerServiceKey] = token

	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "Query", err, query, method, bodyJSON)
	}

	urlc := a.url + "/query/" + query
	urlc = strings.Replace(urlc, "//query", "/query", 1)

	// если в запросе / - значит пробрасываем запрос сразу на апи
	if strings.Contains(query, "/") {
		urlc = a.url + "/" + query
		urlc = strings.Replace(urlc, a.url+"//", a.url+"/", 1)
	}

	res, err := lib.Curl(ctx, method, urlc, bodyJSON, nil, handlers, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return fmt.Sprint(res), err
}

func (a *api) objGet(ctx context.Context, uids string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
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

// ObjAttrUpdate изменение значения аттрибута объекта
func (a *api) objAttrUpdate(ctx context.Context, uid, name, value, src, editor string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
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
	handlers[headerServiceKey] = token
	if a.observeLog {
		defer a.observeLogger(ctx, time.Now(), "ObjCreate", err, urlc, bodymap)
	}

	body, _ := json.Marshal(bodymap)
	urlc = a.url + "/objs?format=json"
	urlc = strings.Replace(urlc, "//objs", "/objs", 1)

	_, err = lib.Curl(ctx, "POST", urlc, string(body), &result, map[string]string{}, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

func (a *api) objDelete(ctx context.Context, uids string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	var urlc string
	token, err := lib.GenXServiceKey(a.domain, []byte(a.projectKey), tokenInterval)
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

	_, err = lib.Curl(ctx, "JSONTOPOST", urlc, string(payloadJson), &result, map[string]string{}, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

func (a *api) observeLogger(ctx context.Context, start time.Time, method string, err error, arguments ...interface{}) {
	logger.Info(ctx, "timing api query",
		zap.String("method", method),
		zap.Float64("timing", time.Since(start).Seconds()),
		zap.String("arguments", fmt.Sprint(arguments)),
		zap.Error(err),
	)
}
