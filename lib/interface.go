package app_lib

import (
	"context"
	"html/template"
	"net/http"
	"sync"

	"git.lowcodeplatform.net/fabric/models"
)

type App interface {
	UrlAPI() string
	UrlORM() string

	// Config
	ConfigGet(key string) (value string)
	ConfigSet(key, value string) (err error)
	ConfigParams() map[string]string

	// Handlers
	TIndex(w http.ResponseWriter, r *http.Request, Config map[string]string) template.HTML
	TBlock(r *http.Request, block models.Data, Config map[string]string) template.HTML
	PIndex(w http.ResponseWriter, r *http.Request)
	ProxyPing(w http.ResponseWriter, r *http.Request)
	BPage(r *http.Request, blockSrc string, objPage models.ResponseData, values map[string]interface{}) (res string, err error)
	GetBlock(w http.ResponseWriter, r *http.Request)

	// Function
	hash(str string) string
	CreateFile(path string)
	isError(err error) bool
	WriteFile(path string, data []byte)
	Curl(method, urlc, bodyJSON string, response interface{}, cookies []*http.Cookie) (result interface{}, err error)
	ModuleBuild(block models.Data, r *http.Request, page models.Data, values map[string]interface{}, enableCache bool) (result ModuleResult)
	ModuleBuildParallel(ctxM context.Context, p models.Data, r *http.Request, page models.Data, values map[string]interface{}, enableCache bool, buildChan chan ModuleResult, wg *sync.WaitGroup)
	ErrorModuleBuild(stat map[string]interface{}, buildChan chan ModuleResult, timerRun interface{}, errT error)
	QueryWorker(queryUID, dataname string, source []map[string]string, r *http.Request) (result interface{}, err error)
	ErrorPage(err interface{}, w http.ResponseWriter, r *http.Request)
	ModuleError(err interface{}, r *http.Request) template.HTML
	GUIQuery(tquery string, r *http.Request) (returnResp Response, err error)

	// Cache
	SetCahceKey(r *http.Request, p models.Data) (key, keyParam string)
	СacheGet(key string, block models.Data, r *http.Request, page models.Data, values map[string]interface{}, url string) (string, bool)
	CacheSet(key string, block models.Data, page models.Data, value, url string) bool
	cacheUpdate(key string, block models.Data, r *http.Request, page models.Data, values map[string]interface{}, url string)
	refreshTime(options models.Data) int

	// DogFunc
	TplValue(v map[string]interface{}, arg []string) (result string)
	ConfigValue(arg []string) (result string)
	SplitIndex(arg []string) (result string)
	Time(arg []string) (result string)
	TimeFormat(arg []string) (result string)
	FuncURL(r *http.Request, arg []string) (result string)
	Path(d []models.Data, arg []string) (result string)
	DReplace(arg []string) (result string)
	UserObj(r *http.Request, arg []string) (result string)
	UserProfile(r *http.Request, arg []string) (result string)
	UserRole(r *http.Request, arg []string) (result string)
	Cookie(r *http.Request, arg []string) (result string)
	Obj(data []models.Data, arg []string) (result string)
	FieldValue(data []models.Data, arg []string) (result string)
	FieldSrc(data []models.Data, arg []string) (result string)
	FieldSplit(data []models.Data, arg []string) (result string)
	DateModify(arg []string) (result string)
	Sendmail(arg []string) (result string)
	Query(r *http.Request, arg []string) (result interface{}, err error)
	DogParse(p string, r *http.Request, queryData *[]models.Data, values map[string]interface{}) (result string)
}
