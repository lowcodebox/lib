package models

import "strings"

type Data struct {
	Uid        string               `json:"uid"`
	Id         string               `json:"id"`
	Source     string               `json:"source"`
	Parent     string               `json:"parent"`
	Type       string               `json:"type"`
	Title      string               `json:"title"`
	Rev        string               `json:"rev"`
	Copies     string               `json:"copies"`
	Attributes map[string]Attribute `json:"attributes"`
}

type Attribute struct {
	Value  string `json:"value"`
	Src    string `json:"src"`
	Tpls   string `json:"tpls"`
	Status string `json:"status"`
	Rev    string `json:"rev"`
	Editor string `json:"editor"`
}

type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Status  RestStatus  `json:"status,omitempty"`
	Metrics Metrics     `json:"metrics,omitempty"`
}

type ResponseData struct {
	Data    []Data      `json:"data"`
	Res     interface{} `json:"res"`
	Status  RestStatus  `json:"status"`
	Metrics Metrics     `json:"metrics"`
}

type Metrics struct {
	ResultSize    int    `json:"result_size,omitempty"`
	ResultCount   int    `json:"result_count,omitempty"`
	ResultOffset  int    `json:"result_offset,omitempty"`
	ResultLimit   int    `json:"result_limit,omitempty"`
	ResultPage    int    `json:"result_page,omitempty"`
	TimeExecution string `json:"time_execution,omitempty"`
	TimeQuery     string `json:"time_query,omitempty"`

	PageLast    int   `json:"page_last,omitempty"`
	PageCurrent int   `json:"page_current,omitempty"`
	PageList    []int `json:"page_list,omitempty"`
	PageFrom    int   `json:"page_from,omitempty"`
	PageTo      int   `json:"page_to,omitempty"`
}

// Attr возвращаем необходимый значение атрибута для объекта если он есть, инае пусто
// а также из заголовка объекта
func (p *Data) Attr(name, element string) (result string, found bool) {

	if _, found := p.Attributes[name]; found {

		// фикс для тех объектов, на которых добавлено скрытое поле Uid
		if name == "uid" {
			return p.Uid, true
		}

		switch element {
		case "src":
			return p.Attributes[name].Src, true
		case "value":
			return p.Attributes[name].Value, true
		case "tpls":
			return p.Attributes[name].Tpls, true
		case "rev":
			return p.Attributes[name].Rev, true
		case "status":
			return p.Attributes[name].Status, true
		case "uid":
			return p.Uid, true
		case "source":
			return p.Source, true
		case "id":
			return p.Id, true
		case "title":
			return p.Title, true
		case "type":
			return p.Type, true
		}
	} else {
		switch element {
		case "uid":
			return p.Uid, true
		case "source":
			return p.Source, true
		case "rev":
			return p.Rev, true
		case "id":
			return p.Id, true
		case "title":
			return p.Title, true
		case "type":
			return p.Type, true
		}
	}
	return "", false
}

// AttrSet заменяем значение аттрибутов в объекте профиля
func (p *Data) AttrSet(name, element, value string) bool {
	g := Attribute{}

	for k, v := range p.Attributes {
		if k == name {
			g = v
		}
	}

	switch element {
	case "src":
		g.Src = value
	case "value":
		g.Value = value
	case "tpls":
		g.Tpls = value
	case "rev":
		g.Rev = value
	case "status":
		g.Status = value
	}

	f := p.Attributes

	for k, _ := range f {
		if k == name {
			f[k] = g
			return true
		}
	}

	return false
}

func (p *Data) HasError() bool {
	return p.Type == "error"
}

// RemoveData удаляем элемент из слайса
func (p *ResponseData) RemoveData(i int) bool {

	if i < len(p.Data) {
		p.Data = append(p.Data[:i], p.Data[i+1:]...)
	} else {
		//log.Warning("Error! Position invalid (", i, ")")
		return false
	}

	return true
}

// FilterRole применяем ограничения доступа для объектов типа ResponseData
// фильтруем массив данных
// если непустое поле access_read, значит назначены права, а следовательно проверяем право просмотра для роли пользователя
// также возвращаем
func (p *ResponseData) FilterRole(role string) {
	sliceData := p.Data

	for i := len(sliceData) - 1; i >= 0; i-- {
		v := sliceData[i]
		attr_read, _ := v.Attr("access_read", "src")
		attr_write, _ := v.Attr("attr_write", "src")
		attr_delete, _ := v.Attr("attr_delete", "src")
		attr_admin, _ := v.Attr("attr_admin", "src")

		if (!strings.Contains(attr_read, role) || attr_read == "") &&
			(!strings.Contains(attr_write, role) || attr_write == "") &&
			(!strings.Contains(attr_delete, role) || attr_delete == "") &&
			(!strings.Contains(attr_admin, role) || attr_admin == "") {
			p.RemoveData(i)
		}
	}

	return
}
