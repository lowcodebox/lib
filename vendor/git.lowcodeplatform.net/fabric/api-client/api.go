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
)

const headerRequestId = "X-Request-Id"

type api struct {
	url string
}

type Api interface {
	Obj
}

type Obj interface {
	ObjGet(ctx context.Context, uids string) (result models.ResponseData, err error)
	ObjCreate(ctx context.Context, bodymap map[string]string) (result models.ResponseData, err error)
	ObjAttrUpdate(ctx context.Context, uid, name, value, src, editor string) (result models.ResponseData, err error)
	LinkGet(ctx context.Context, tpl, obj, mode, short string) (result models.ResponseData, err error)
	Query(ctx context.Context, query, method, bodyJSON string) (result string, err error)
	Element(ctx context.Context, action, body string) (result models.ResponseData, err error)
}

// Query результат выводим в объект как при вызове Curl
func (o *api) Query(ctx context.Context, query, method, bodyJSON string) (result string, err error) {
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)

	urlc := o.url + "/query/" + query
	urlc = strings.Replace(urlc, "//query", "/query", 1)

	// если в запросе / - значит пробрасываем запрос сразу на апи
	if strings.Contains(query, "/") {
		urlc = o.url + "/" + query
		urlc = strings.Replace(urlc, o.url+"//", o.url+"/", 1)
	}

	res, err := lib.Curl(method, urlc, bodyJSON, nil, map[string]string{}, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return fmt.Sprint(res), err
}

func (o *api) ObjGet(ctx context.Context, uids string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)

	urlc := o.url + "/objs/" + uids
	urlc = strings.Replace(urlc, o.url+"//objs/", o.url+"/objs/", 1)

	_, err = lib.Curl("GET", urlc, "", &result, map[string]string{}, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

func (o *api) LinkGet(ctx context.Context, tpl, obj, mode, short string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)

	urlc := o.url + "/link/get?source=" + tpl + "&mode=" + mode + "&obj=" + obj + "&short=" + short
	urlc = strings.Replace(urlc, "//link", "/link", 1)

	_, err = lib.Curl("GET", urlc, "", &result, map[string]string{}, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, urlc)
	}

	return result, err
}

// ObjAttrUpdate изменение значения аттрибута объекта
func (a *api) ObjAttrUpdate(ctx context.Context, uid, name, value, src, editor string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)

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
func (a *api) Element(ctx context.Context, action, body string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)

	// получаем поля шаблона
	if action == "elements" || action == "all" {
		_, err = lib.Curl("GET", a.url+"/element/"+body, "", &result, map[string]string{}, nil)
		if err != nil {
			err = fmt.Errorf("%s (url: %s)", err, a.url+"/element/"+body)
		}
		return result, err
	}

	_, err = curl.NewRequestDefault().Method("POST").Payload(body).MapToObj(&result).Url(a.url + "/element/" + action + "?format=json").Do(nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, a.url+"/element/"+action+"?format=json")
	}

	return result, err
}

func (a *api) ObjCreate(ctx context.Context, bodymap map[string]string) (result models.ResponseData, err error) {
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(ctx)

	body, _ := json.Marshal(bodymap)
	_, err = lib.Curl("POST", a.url+"/objs?format=json", string(body), &result, map[string]string{}, nil)
	if err != nil {
		err = fmt.Errorf("%s (url: %s)", err, a.url+"/objs?format=json")
	}

	return result, err
}

func New(url string) Api {
	return &api{
		url,
	}
}
