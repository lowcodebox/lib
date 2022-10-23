package models

type Data struct {
	Uid        		string               `json:"uid"`
	Id         		string               `json:"id"`
	Source     		string               `json:"source"`
	Parent     		string               `json:"parent"`
	Type       		string               `json:"type"`
	Title      		string               `json:"title"`
	Rev        		string               `json:"rev"`
	Сopies			string 				 `json:"copies"`
	Attributes 		map[string]Attribute `json:"attributes"`
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
	Data   	interface{} 	`json:"data"`
	Status 	RestStatus    	`json:"status"`
	Metrics Metrics 		`json:"metrics"`
}

type ResponseData struct {
	Data      []Data        `json:"data"`
	Res   	  interface{} 	`json:"res"`
	Status    RestStatus    `json:"status"`
	Metrics   Metrics 		`json:"metrics"`
}

type Metrics struct {
	ResultSize     	int `json:"result_size"`
	ResultCount     int `json:"result_count"`
	ResultOffset    int `json:"result_offset"`
	ResultLimit     int `json:"result_limit"`
	ResultPage 		int `json:"result_page"`
	TimeExecution   string `json:"time_execution"`
	TimeQuery   	string `json:"time_query"`

	PageLast		int `json:"page_last"`
	PageCurrent		int `json:"page_current"`
	PageList		[]int `json:"page_list"`
	PageFrom		int `json:"page_from"`
	PageTo			int `json:"page_to"`
}

// возвращаем необходимый значение атрибута для объекта если он есть, инае пусто
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
		switch name {
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
	}
	return "", false
}

// заменяем значение аттрибутов в объекте профиля
func (p *Data) AttrSet(name, element, value string) bool  {
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

// удаляем элемент из слайса
func (p *ResponseData) RemoveData(i int) bool {

	if (i < len(p.Data)){
		p.Data = append(p.Data[:i], p.Data[i+1:]...)
	} else {
		//log.Warning("Error! Position invalid (", i, ")")
		return false
	}

	return true
}