package app_lib

import (
	"html/template"
	"net/http"
	"sync"

	"git.lowcodeplatform.net/fabric/lib"
	"github.com/restream/reindexer"
)

type app struct {
	logger         *lib.Log
	serviceMetrics lib.ServiceMetric
	urlORM         string `json:"url_orm"`
	urlAPI         string `json:"url_api"`
	db             *reindexer.Reindexer
	pageSize       int
	status         string
	count          string
	config         Cfg
	vfs            lib.Vfs
}

type Cfg struct {
	payload map[string]string
	mx      sync.Mutex
}

var t *template.Template
var result template.HTML
var debugMode = true
var FlagParallel = true // флаг генерации блоков в параллельном режиме
var Metric template.HTML

var Domain, Title, UidAPP, ClientPath, UidPrecess, LogsDir, LogsLevel string
var UrlAPI, UrlORM string
var ReplicasService int

// тип ответа, который сервис отдает прокси при периодическом опросе (ping-е)
type Pong struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Status   string `json:"status"`
	Port     int    `json:"port"`
	Pid      string `json:"pid"`
	State    string `json:"state"`
	Replicas int    `json:"replicas"`
}

type ModuleResult struct {
	id     string
	result template.HTML
	stat   map[string]interface{}
	err    error
}

type ProfileData struct {
	Hash           string
	Email          string
	Uid            string
	First_name     string
	Last_name      string
	Photo          string
	Age            string
	City           string
	Country        string
	Status         string // - src поля Status в профиле (иногда необходимо для доп.фильтрации)
	Raw            []Data // объект пользователя (нужен при сборки проекта для данного юзера при добавлении прав на базу)
	Tables         []Data
	Roles          []Data
	Homepage       string
	Maket          string
	UpdateFlag     bool
	UpdateData     []Data
	CurrentRole    Data
	CurrentProfile Data
	Navigator      []*Items
}

type Items struct {
	Title        string   `json:"title"`
	ExtentedLink string   `json:"extentedLink"`
	Uid          string   `json:"uid"`
	Source       string   `json:"source"`
	Icon         string   `json:"icon"`
	Leader       string   `json:"leader"`
	Order        string   `json:"order"`
	Type         string   `json:"type"`
	Preview      string   `json:"preview"`
	Url          string   `json:"url"`
	Sub          []string `json:"sub"`
	Incl         []*Items `json:"incl"`
	Class        string   `json:"class"`
}

type Request struct {
	Data []interface{} `json:"data"`
}

type Response struct {
	Data    interface{} `json:"data"`
	Status  RestStatus  `json:"status"`
	Metrics Metrics     `json:"metrics"`
}

type Metrics struct {
	ResultSize    int    `json:"result_size"`
	ResultCount   int    `json:"result_count"`
	ResultOffset  int    `json:"result_offset"`
	ResultLimit   int    `json:"result_limit"`
	ResultPage    int    `json:"result_page"`
	TimeExecution string `json:"time_execution"`
	TimeQuery     string `json:"time_query"`

	PageLast    int   `json:"page_last"`
	PageCurrent int   `json:"page_current"`
	PageList    []int `json:"page_list"`
	PageFrom    int   `json:"page_from"`
	PageTo      int   `json:"page_to"`
}

type RestStatus struct {
	Description string `json:"description"`
	Status      int    `json:"status"`
	Code        string `json:"code"`
	Error       error  `json:"error"`
}

type ResponseData struct {
	Data    []Data      `json:"data"`
	Res     interface{} `json:"res"`
	Status  RestStatus  `json:"status"`
	Metrics Metrics     `json:"metrics"`
}

type Attribute struct {
	Value  string `json:"value"`
	Src    string `json:"src"`
	Tpls   string `json:"tpls"`
	Status string `json:"status"`
	Rev    string `json:"rev"`
	Editor string `json:"editor"`
}

type Data struct {
	Uid        string               `json:"uid"`
	Id         string               `json:"id"`
	Source     string               `json:"source"`
	Parent     string               `json:"parent"`
	Type       string               `json:"type"`
	Title      string               `json:"title"`
	Rev        string               `json:"rev"`
	Сopies     string               `json:"copies"`
	Attributes map[string]Attribute `json:"attributes"`
}

type DataTree struct {
	Uid        string               `json:"uid"`
	Id         string               `json:"id"`
	Source     string               `json:"source"`
	Parent     string               `json:"parent"`
	Type       string               `json:"type"`
	Title      string               `json:"title"`
	Rev        string               `json:"rev"`
	Сopies     string               `json:"copies"`
	Attributes map[string]Attribute `json:"attributes"`
	Sub        []string             `json:"sub"`
	Incl       []*DataTree          `json:"incl"`
}

type DataTreeOut struct {
	Uid        string               `json:"uid"`
	Id         string               `json:"id"`
	Source     string               `json:"source"`
	Parent     string               `json:"parent"`
	Type       string               `json:"type"`
	Title      string               `json:"title"`
	Rev        string               `json:"rev"`
	Сopies     string               `json:"copies"`
	Attributes map[string]Attribute `json:"attributes"`
	Sub        []string             `json:"sub"`
	Incl       []DataTree           `json:"incl"`
}

type ValueCache struct {
	Uid      string   `reindex:"uid,,pk"`
	Link     []string `reindex:"link"`
	Value    string   `reindex:"value"`
	Deadtime string   `reindex:"deadtime"`
	Status   string   `reindex:"status"`
	Url      string   `reindex:"url"`
}

// элемент конфигурации - поле
type Element struct {
	Type   string
	Source interface{}
}

// элемент конфигурации - поле
type ErrorForm struct {
	Err interface{}
	R   interface{}
}

type Page struct {
	Title   string                 `json:"title"`
	Prefix  interface{}            `json:"prefix"`
	Request interface{}            `json:"request"`
	Domain  string                 `json:"domain"`
	Blocks  map[string]interface{} `json:"blocks"`
	Data    interface{}            `json:"data"`
	Shema   interface{}            `json:"shema"`
	CSS     []string               `json:"css"`
	JS      []string               `json:"js"`
	JSH     []string               `json:"jsh"`
	CSSC    []string               `json:"cssc"`
	JSC     []string               `json:"jsc"`
	Metric  template.HTML          `json:"metric"`
	Stat    []interface{}          `json:"stat"`
}

type Block struct {
	Page             interface{}            `json:"page"`
	Data             interface{}            `json:"data"`
	Configuration    interface{}            `json:"configuration"`
	ConfigurationRaw string                 `json:"configuration_raw"`
	Value            map[string]interface{} `json:"value"`
	CSS              []string               `json:"css"`
	JS               []string               `json:"js"`
	Metric           template.HTML          `json:"metric"`
	Request          *http.Request
	mx               sync.Mutex
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

func (p *Block) Set(key, value string) {
	p.Value[key] = value
	return
}

func (p *Block) Get(key string) interface{} {
	return p.Value[key]
}

func (p *Block) CSSPath() {
	return
}

func New(logger *lib.Log, metric lib.ServiceMetric, urlORM, urlAPI string, db *reindexer.Reindexer, status, count string, config map[string]string, vfs lib.Vfs) App {

	// добавляем карту функций FuncMap функциями из библиотеки github.com/Masterminds/sprig
	// только те, которые не описаны в FuncMap самостоятельно
	for k, v := range FuncMapS {
		if _, found := FuncMap[k]; !found {
			FuncMap[k] = v
		}
	}
	conf := Cfg{
		payload: config,
	}

	return &app{
		logger:         logger,
		serviceMetrics: metric,
		urlAPI:         urlAPI,
		urlORM:         urlORM,
		db:             db,
		status:         status,
		count:          count,
		config:         conf,
		vfs:            vfs,
	}
}
