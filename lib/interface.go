package app_lib

import (
	"net/http"
	"html/template"
	"sync"
	"context"
)

type App interface {

	UrlAPI() string
	UrlORM() string
	
	// Config
	ConfigGet(key string) (value string)
	ConfigSet(key, value string) (err error)
	ConfigParams() (map[string]string)
	
	// Handlers
	TIndex(w http.ResponseWriter, r *http.Request, Config map[string]string) template.HTML
	TBlock(r *http.Request, block Data, Config map[string]string) template.HTML
	PIndex(w http.ResponseWriter, r *http.Request)
	ProxyPing(w http.ResponseWriter, r *http.Request)
	BPage(r *http.Request, blockSrc string, objPage ResponseData, values map[string]interface{}) string
	GetBlock(w http.ResponseWriter, r *http.Request)
	
	// Function
	hash(str string) string
	CreateFile(path string)
	isError(err error) bool
	WriteFile(path string, data []byte)
	Curl(method, urlc, bodyJSON string, response interface{}, cookies []*http.Cookie) (result interface{}, err error)
	ModuleBuild(block Data, r *http.Request, page Data, values map[string]interface{}, enableCache bool) (result ModuleResult)
	ModuleBuildParallel(ctxM context.Context, p Data, r *http.Request, page Data, values map[string]interface{}, enableCache bool, buildChan chan ModuleResult, wg *sync.WaitGroup)
	ErrorModuleBuild(stat map[string]interface{}, buildChan chan ModuleResult, timerRun interface{}, errT error)
	QueryWorker(queryUID, dataname string, source[]map[string]string, r *http.Request) interface{}
	ErrorPage(err interface{}, w http.ResponseWriter, r *http.Request)
	ModuleError(err interface{}, r *http.Request) template.HTML
	GUIQuery(tquery string, r *http.Request) Response
	
	// Cache
	SetCahceKey(r *http.Request, p Data) (key, keyParam string)
	СacheGet(key string, block Data, r *http.Request, page Data, values map[string]interface{}, url string) (string, bool)
	CacheSet(key string, block Data, page Data, value, url string) bool
	cacheUpdate(key string, block Data, r *http.Request, page Data, values map[string]interface{}, url string)
	refreshTime(options Data) int
	
	// DogFunc
	TplValue(v map[string]interface{}, arg []string) (result string)
	ConfigValue(arg []string) (result string)
	SplitIndex(arg []string) (result string)
	Time(arg []string) (result string)
	TimeFormat(arg []string) (result string)
	FuncURL(r *http.Request, arg []string) (result string)
	Path(d []Data, arg []string) (result string)
	DReplace(arg []string) (result string)
	UserObj(r *http.Request, arg []string) (result string)
	UserProfile(r *http.Request, arg []string) (result string)
	UserRole(r *http.Request, arg []string) (result string)
	Cookie(r *http.Request, arg []string) (result string)
	Obj(data []Data, arg []string) (result string)
	FieldValue(data []Data, arg []string) (result string)
	FieldSrc(data []Data, arg []string) (result string)
	FieldSplit(data []Data, arg []string) (result string)
	DateModify(arg []string) (result string)
	Sendmail(arg []string) (result string)
	Query(r *http.Request, arg []string) (result interface{})
	DogParse(p string, r *http.Request, queryData *[]Data, values map[string]interface{}) (result string)
}
