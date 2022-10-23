package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
)

type api struct {
	url    string
	logger lib.Log
	metric lib.ServiceMetric
}

type Api interface {
	Obj
}

type Obj interface {
	ObjGet(uids string) (result models.ResponseData, err error)
	ObjCreate(bodymap map[string]string) (result models.ResponseData, err error)
	ObjAttrUpdate(uid, name, value, src, editor string) (result models.ResponseData, err error)
	LinkGet(tpl, obj, mode, short string) (result models.ResponseData, err error)
	Query(query, method, bodyJSON string) (result string, err error)
}

// результат выводим в объект как при вызове Curl
func (o *api) Query(query, method, bodyJSON string) (result string, err error) {
	urlc := o.url + "/query/" + query
	urlc = strings.Replace(urlc, "//query", "/query", 1)

	res, err := lib.Curl(method, urlc, bodyJSON, nil, map[string]string{}, nil)
	return fmt.Sprint(res), err
}

func (o *api) ObjGet(uids string) (result models.ResponseData, err error) {
	urlc := o.url + "/query/obj?obj=" + uids
	urlc = strings.Replace(urlc, "//query", "/query", 1)

	_, err = lib.Curl("GET", urlc, "", &result, map[string]string{}, nil)
	return result, err
}

func (o *api) LinkGet(tpl, obj, mode, short string) (result models.ResponseData, err error) {
	urlc := o.url + "/link/get?source=" + tpl + "&mode=" + mode + "&obj=" + obj + "&short=" + short
	urlc = strings.Replace(urlc, "//link", "/link", 1)

	_, err = lib.Curl("GET", urlc, "", &result, map[string]string{}, nil)

	return result, err
}

// изменение значения аттрибута объекта
func (a *api) ObjAttrUpdate(uid, name, value, src, editor string) (result models.ResponseData, err error) {

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
	result, err = a.Element("update", string(dataJ))

	return result, err
}

// Element
// TODO ПЕРЕДЕЛАТЬ на понятные пути в ORM
// сделано так для совместимости со старой версией GUI
func (a *api) Element(action, body string) (result models.ResponseData, err error) {
	_, err = lib.Curl("POST", a.url+"/element/"+action, body, &result, map[string]string{}, nil)

	return result, err
}

func (a *api) ObjCreate(bodymap map[string]string) (result models.ResponseData, err error) {
	body, _ := json.Marshal(bodymap)
	_, err = lib.Curl("POST", a.url+"/objs?format=json", string(body), &result, map[string]string{}, nil)

	return result, err
}

func New(url string, logger lib.Log, metric lib.ServiceMetric) Api {
	return &api{
		url,
		logger,
		metric,
	}
}
