package app_lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
	"git.edtech.vm.prod-6.cloud.el/packages/logger"
	"github.com/gorilla/mux"
)

// таймаут срабатывания завершения обработки модулей (через отмену контектста и таймаут внешних запросов)
var timeoutBlockGen = 10 * time.Second

// ответ на запрос прокси
func (c *app) ProxyPing(w http.ResponseWriter, r *http.Request) {

	pp := strings.Split(c.ConfigGet("domain"), "/")
	pg, _ := strconv.Atoi(c.ConfigGet("PortAPP"))
	//runfile := c.State["runfile"]

	name := "ru"
	version := "ru"
	if len(pp) == 1 {
		name = pp[0]
	}

	if len(pp) == 2 {
		name = pp[0]
		version = pp[1]
	}

	pid := strconv.Itoa(os.Getpid()) + ":" + c.ConfigGet("data-uid")
	state, _ := json.Marshal(c.serviceMetrics.Get())

	var pong = []Pong{
		{name, version, "run", pg, pid, string(state), ReplicasService},
	}

	// заменяем переменную домена на правильный формат для использования в Value.Prefix при генерации страницы
	c.ConfigSet("domain", name+"/"+version)
	c.ConfigSet("client_path", "/"+name+"/"+version)

	res, _ := json.Marshal(pong)

	w.WriteHeader(200)
	w.Write([]byte(res))
}

// PIndex Собираем страницу (параметры из конфига) и пишем в w.Write
func (c *app) PIndex(w http.ResponseWriter, r *http.Request) {
	var objPages, objPage models.ResponseData
	var res string
	var err error
	vars := mux.Vars(r)

	defer func() {
		if err != nil {
			w.WriteHeader(501)
			w.Write([]byte(fmt.Sprint(err)))
			return
		}
	}()

	// указатель на профиль текущего пользователя
	ctx := r.Context()
	var profile ProfileData
	profileRaw := ctx.Value("UserRaw")
	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)

	// получаем параметры из файла конфигурации (лежат в переменной Application)
	page := vars["page"]
	path_template := c.ConfigGet("path_templates")
	tpl_app_blocks_pointsrc := c.ConfigGet("tpl_app_blocks_pointsrc")
	tpl_app_pages_pointsrc := c.ConfigGet("tpl_app_pages_pointsrc")
	data_source := c.ConfigGet("data-source")
	title := c.ConfigGet("title")
	Domain = c.ConfigGet("domain")
	ClientPath = c.ConfigGet("client_path")
	//if ClientPath == "" {
	//	ClientPath = Domain
	//}

	// ПЕРЕДЕЛАТЬ или на кеширование страниц и на доп.проверку
	if page == "" {
		// получаем все страницы текущего приложения
		//c.Curl("GET", "_link?obj="+data_source+"&source="+tpl_app_pages_pointsrc+"&mode=out", "", &objPages, r.Cookies())
		objPages, err = c.api.LinkGet(r.Context(), tpl_app_pages_pointsrc, data_source, "out", "")
		if err != nil {
			err = fmt.Errorf("error pages current user. err: %s", err)
			return
		}
		for _, v := range objPages.Data {
			if def, _ := v.Attr("default", "value"); def == "checked" {
				page = v.Uid
			}
		}
	}
	if page == "" {
		ff, _ := json.Marshal(objPages)
		w.WriteHeader(403)
		w.Write([]byte("Error: not default page (" + fmt.Sprint(ff) + ")"))
		return
	}

	// запрос объекта страницы
	//_, err = c.Curl("GET", "_objs/"+page, "", &objPage, r.Cookies())
	objPages, err = c.api.ObjGet(r.Context(), page)
	if err != nil {
		err = fmt.Errorf("error get page. err: %s", err)
		return
	}

	// ФИКС! иногда в разных приложениях называют одинаково страницы.
	// удаляем из объекта objPage значения не текущего приложения
	if len(objPage.Data) > 1 {
		for k, v := range objPage.Data {
			app, _ := v.Attr("app", "src")
			if app != c.ConfigGet("data-uid") {
				objPage.RemoveData(k)
			}
		}
	}

	// формируем значение переменных, переданных в страницу
	values := map[string]interface{}{}

	//pp := strings.Split(Domain, "/")
	//if len(pp) == 1 {
	//	ClientPath = Domain + "/" + "gui"
	//}

	//jsonRequest, _ := json.Marshal(r)
	// values["Request"] = string(jsonRequest)

	values["Prefix"] = ClientPath + path_template
	values["Domain"] = Domain
	values["Path"] = ClientPath
	values["CDN"] = ""
	values["Title"] = title
	values["URL"] = r.URL.Query().Encode()
	values["Referer"] = r.Referer()
	values["RequestURI"] = r.RequestURI
	values["Profile"] = profile
	values["ProxyPath"] = c.ConfigGet("ProxyPointsrc")

	res, err = c.BPage(r, tpl_app_blocks_pointsrc, objPage, values)
	if err != nil {

	}

	w.WriteHeader(200)
	w.Write([]byte(res))
}

// TIndex возвращаем сформированную страницу в template.HTML (для cockpit-a и dashboard)
func (c *app) TIndex(w http.ResponseWriter, r *http.Request, Config map[string]string) template.HTML {

	var objPage, objApp models.ResponseData
	vars := mux.Vars(r)
	page := vars["obj"] // ид-страницы передается через переменную obj

	// указатель на профиль текущего пользователя
	ctx := r.Context()
	var profile ProfileData
	profileRaw := ctx.Value("UserRaw")
	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)

	//var profile ProfileData
	////pp := fmt.Sprint(reflect.ValueOf(dlink).Elem())
	//bb, err := json.Marshal(dlink)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//json.Unmarshal(bb, &profile)
	//
	//fmt.Println("\ndlink\n\n", dlink)
	//fmt.Println("\nprofile\n\n", profile)

	// можем задать также через &page=страница
	if page == "" {
		if r.FormValue("page") != "" {
			page = r.FormValue("page")
		}

		if page == "" {
			return ""
		}
	}

	// заменяем значения при вызове ф-ции из GUI ибо они пустые, ведь приложение полностью не инициализировано через конфиг

	//JRequest 	:= Config["request"] // ? что это? разобраться!
	Domain = c.ConfigGet("domain")
	ClientPath = c.ConfigGet("client_path")
	Title = c.ConfigGet("title")
	Metric = template.HTML(c.ConfigGet("metric"))

	if page == "" {
		return template.HTML("Error: Not id page")
	}

	// запрос объекта страницы
	//c.Curl("GET", "_objs/"+page, "", &objPage, r.Cookies())
	objPage, err := c.api.ObjGet(r.Context(), page)
	if err != nil {
		err = fmt.Errorf("error get page (%s). err: %s", page, err)
		return template.HTML(fmt.Sprintf("%s", err))
	}
	if &objPage == nil {
		return template.HTML("Error: Not found page-object.") // если не найден объект страницы
	}
	if len(objPage.Data) == 0 {
		return template.HTML("Error: Not found page-object.") // если не найден объект страницы
	}

	// Uid-приложения
	appUid, found := objPage.Data[0].Attr("app", "src")
	if !found {
		return template.HTML("Error: Not selected application from this page:" + page)
	}

	// запрос объекта приложения
	//c.Curl("GET", "_objs/"+appUid, "", &objApp, r.Cookies())
	objApp, err = c.api.ObjGet(r.Context(), appUid)
	if err != nil {
		err = fmt.Errorf("error get obj application (%s). err: %s", appUid, err)
		return template.HTML(fmt.Sprintf("%s", err))
	}
	if &objApp == nil {
		return template.HTML("Error: Not found application-object.") // если не найден объект приложения
	}
	if len(objApp.Data) == 0 {
		return template.HTML("Error: Not found application-object.") // если не найден объект страницы
	}

	//fmt.Println("objApp: ", objApp)

	// получаем значения аттрибутов для данного приложения
	path_template, found := objApp.Data[0].Attr("path_templates", "value")
	if !found {
		return template.HTML("Error: Not selected path_templates from this application.")
	}

	// получаем значения аттрибутов для данного приложения
	tpl_app_blocks_pointsrc, found := objApp.Data[0].Attr("tpl_app_blocks", "src")
	if !found {
		return template.HTML("Error: Not selected tpl_app_blocks from this application.")
	}

	//pp := strings.Split(Domain, "/")
	//if len(pp) == 1 {
	//	ClientPath = Domain + "/" + "gui"
	//}

	// получили значение Request в json - возвращаем в http.Request
	//var PageRequest *http.Request
	//json.Unmarshal([]byte(JRequest), &PageRequest)

	// формируем значение переменных, переданных в страницу
	values := map[string]interface{}{}
	values["Prefix"] = ClientPath + path_template
	values["Domain"] = Domain
	values["Path"] = ClientPath
	values["CDN"] = ""
	values["Title"] = Title
	values["URL"] = r.URL.Query().Encode()
	values["Referer"] = r.Referer()
	values["RequestURI"] = r.RequestURI
	values["Profile"] = profile
	values["ProxyPath"] = c.ConfigGet("ProxyPointsrc")

	res, err := c.BPage(r, tpl_app_blocks_pointsrc, objPage, values)
	if err != nil {
		return template.HTML(fmt.Sprint(err))
	}

	return template.HTML(res)
}

// BPage собираем страницу
func (l *app) BPage(r *http.Request, blockSrc string, objPage models.ResponseData, values map[string]interface{}) (res string, err error) {
	var objMaket, objBlocks models.ResponseData
	moduleResult := ModuleResult{}
	statModule := map[string]interface{}{}

	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(r.Context())

	// флаг режима генерации модулей (последовательно/параллельно)

	p := &Page{}
	p.Title = Title
	p.Domain = Domain
	p.Metric = Metric
	p.Prefix = values["Prefix"]
	//p.Request 	= values["Request"]
	p.CSS = []string{}
	p.JS = []string{}
	p.JSH = []string{}
	p.CSSC = []string{}
	p.JSC = []string{}
	p.Stat = []interface{}{}
	p.Blocks = map[string]interface{}{}

	if len(objPage.Data) == 0 {
		return "", fmt.Errorf("error: Object page is null")
	}

	pageUID := objPage.Data[0].Uid
	maketUID, _ := objPage.Data[0].Attr("maket", "src")

	// 1.0 проверка на принадлежность страницы текущему проекту
	// ДОДЕЛАТЬ СРОЧНО!!!

	// 2 запрос на объекты блоков страницы
	//urlc := l.urlAPI + "_link?obj=" + pageUID + "&source=" + blockSrc + "&mode=in"
	//urlc = strings.Replace(urlc, "//_link", "/_link", 1)
	//_, err = lib.Curl("GET", urlc, "", &objBlocks, handlers, r.Cookies())

	objBlocks, err = l.api.LinkGet(r.Context(), blockSrc, pageUID, "in", "")
	if err != nil {
		err = fmt.Errorf("error pages current user. err: %s", err)
		return
	}
	if len(objBlocks.Data) == 0 {
		err = fmt.Errorf("error: Objects block %s is null", blockSrc)
		return "", err
	}

	//for _, v := range objBlocks.Data {
	//	fmt.Println("objBlocks: ", v.Title, v.Id)
	//}

	// 3 запрос на объект макета
	//urlc = l.urlAPI + "_objs/" + maketUID
	//urlc = strings.Replace(urlc, "//_objs", "/_objs", 1)
	//_, err = lib.Curl("GET", urlc, "", &objMaket, handlers, r.Cookies())

	objMaket, err = l.api.ObjGet(r.Context(), maketUID)
	if err != nil {
		err = fmt.Errorf("%s (%s)", "Error: Get object Module is failed!", err)
		res = fmt.Sprint(err)
	}
	if &objMaket == nil {
		return "", fmt.Errorf("error: Not found maketUID: %s", maketUID) // если не найден объект
	}
	if len(objMaket.Data) == 0 {
		return "", fmt.Errorf("error: Data maketUID id empty. maketUID: %s", maketUID) // если не найден объект
	}

	// 4 из объекта макета берем путь к шаблону + css и js
	maketFile, _ := objMaket.Data[0].Attr("file", "value")
	maketFileInside, _ := objMaket.Data[0].Attr("_filecontent_file", "value")
	maketCSS, _ := objMaket.Data[0].Attr("css", "value")
	maketJS, _ := objMaket.Data[0].Attr("js", "value")
	maketJSH, _ := objMaket.Data[0].Attr("jsh", "value")
	maketJSC, _ := objMaket.Data[0].Attr("js_custom", "value")
	maketCSSC, _ := objMaket.Data[0].Attr("css_custom", "value")

	// 5 добавляем в объект страницы список файлов css и js
	for _, v := range strings.Split(maketCSS, ";") {
		p.CSS = append(p.CSS, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(maketJS, ";") {
		p.JS = append(p.JS, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(maketJSH, ";") {
		p.JSH = append(p.JSH, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(maketJSC, ";") {
		p.JSC = append(p.JSC, strings.TrimSpace(v))
	}
	for _, v := range strings.Split(maketCSSC, ";") {
		p.CSSC = append(p.CSSC, strings.TrimSpace(v))
	}

	// 3 сохраняем схему
	var i interface{}
	shemaJSON, _ := objPage.Data[0].Attr("shema", "value")
	json.Unmarshal([]byte(shemaJSON), &i)
	if i == nil {
		return "", fmt.Errorf("error! Fail json shema")
	}
	p.Shema = i

	// 4 запускаем сборку модулей (получаем сгенерированный template.HTML без JS и CSS
	// шаблоны рендерятся в каждом модуле отдельно (можно далее хранить в кеше)

	//fmt.Println("Собираем страницу, FlagParallel:", FlagParallel)

	if FlagParallel {
		ctx := context.WithValue(context.Background(), "timeout", timeoutBlockGen)
		ctx, cancel := context.WithCancel(ctx)

		// ПАРАЛЛЕЛЬНО
		wg := &sync.WaitGroup{}
		var buildChan = make(chan ModuleResult, len(objBlocks.Data))

		for _, v := range objBlocks.Data {
			idBlock, _ := v.Attr("id", "value") // название блока

			if strings.Contains(shemaJSON, idBlock) { // наличие этого блока в схеме
				wg.Add(1)
				go l.ModuleBuildParallel(ctx, v, r, objPage.Data[0], values, true, buildChan, wg)
			}
		}

		// ждем завершения интервала и вызываем завершение контекста для запущенных воркеров
		exitTimer := make(chan struct{})
		timerBlockGen := time.NewTimer(timeoutBlockGen)
		flagWG := true
		go func() {
			select {
			case <-timerBlockGen.C:
				flagWG = false
				cancel()
				return
			case <-exitTimer:
				timerBlockGen.Stop()
				return
			}
		}()

		// отменяем ожидание wg при условии, что завершился таймаут и нам не нужны результаты недополученных ModuleBuildParallel
		// wg завершатся сами через defer позже
		if flagWG {
			wg.Wait()
		}
		if timerBlockGen.Stop() {
			exitTimer <- struct{}{}
		}

		close(buildChan)

		for k := range buildChan {
			p.Blocks[k.id] = k.result
			p.Stat = append(p.Stat, k.stat)
		}

	} else {

		// ПОСЛЕДОВАТЕЛЬНО
		for _, v := range objBlocks.Data {

			idBlock, _ := v.Attr("id", "value")       // название блока
			if strings.Contains(shemaJSON, idBlock) { // наличие этого блока в схеме
				moduleResult = l.ModuleBuild(v, r, objPage.Data[0], values, true)

				p.Blocks[v.Id] = moduleResult.result
				statModule = moduleResult.stat

				statModule["id"] = v.Id
				statModule["title"] = v.Title
				p.Stat = append(p.Stat, statModule)
			}
		}

	}

	//fmt.Println("Statistic generate page: ", p.Stat)
	//log.Warning("Time всего: ", time.Since(t1))

	// 5 генерируем страницу, использую шаблон выбранной в объекте страницы, схему
	var c bytes.Buffer

	// СЕКЬЮРНО! Если мы вычитаем текущий путь пользователя, то сможем получить доступ к файлам только текущего проекта
	// иначе необходимо будет авторизоваться и правильный путь (например  /console/gui мы не вычтем)
	// НО ПРОБЛЕМА реиспользования ранее загруженных и настроенных путей к шаблонам.
	//maketFile = strings.Replace(maketFile, Application["client_path"], ".", -1)

	// НЕ СЕКЬЮРНО!
	// вычитаем не текущий client_path а просто две первых секции из адреса к файлу
	// позволяем получить доступ к ранее загруженным путям шаблонов другим пользоватем с другим префиксом
	// ПО-УМОЛЧАНИЮ (для реиспользования модулей и схем)

	// приоритетнее сохраненные в объекте данные
	var dataFile string
	if maketFileInside != "" {
		dataFile = maketFileInside
	} else {
		sliceMake := strings.Split(maketFile, "/")
		maketFile = strings.Join(sliceMake[3:], "/")

		byteFile, _, err := l.vfs.Read(r.Context(), maketFile)
		if err != nil {
			return "", fmt.Errorf("error vfs.Read, maketFile: %s, err: %s", maketFile, err)
		}
		dataFile = string(byteFile)
	}

	// в режиме отладки пересборка шаблонов происходит при каждом запросе
	if debugMode {
		//t = template.Must(template.New(maketFile).Funcs(funcMap).ParseFiles(maketFile))
		//t = template.Must(template.ParseFiles(maketFile))
		tmp := template.New(maketFile)
		t, err = tmp.Parse(dataFile)
		if err != nil {
			return "", fmt.Errorf("error tmp.Parse, maketFile: %s, data: %s, err: %s", maketFile, string(dataFile), err)
		}
		t.Execute(&c, p)
	} else {
		t.ExecuteTemplate(&c, maketFile, p)
	}

	return c.String(), err
}

// GetBlock генерируем один блок - внешний запрос
func (c *app) GetBlock(w http.ResponseWriter, r *http.Request) {
	var objBlock models.ResponseData
	var err error
	var handlers = map[string]string{}
	handlers[headerRequestId] = logger.GetRequestIDCtx(r.Context())

	vars := mux.Vars(r)
	block := vars["block"]
	dataPage := models.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы

	// 3 запрос на объект макета
	//urlc = c.urlAPI + "_objs/" + block
	//urlc = strings.Replace(urlc, "//_objs", "/_objs", 1)
	//_, err = lib.Curl("GET", urlc, "", &objBlock, handlers, r.Cookies())

	objBlock, err = c.api.ObjGet(r.Context(), block)
	if err != nil {
		err = fmt.Errorf("error: Get object is failed %s (%s)", block, err)
		w.Write([]byte(fmt.Sprint(err))) // если не найден объект
		return
	}
	if &objBlock == nil {
		w.Write([]byte("Error: block is not found. block: " + block)) // если не найден объект
		return
	}
	if len(objBlock.Data) == 0 {
		w.Write([]byte("Error: objBlock.Data is empty. block: " + block)) // если не найден объект
		return
	}

	moduleResult := c.ModuleBuild(objBlock.Data[0], r, dataPage, nil, false)

	w.Write([]byte(moduleResult.result))
}

// TBlock генерируем один блок через внутренний запрос - для cocpit-a
func (c *app) TBlock(r *http.Request, block models.Data, Config map[string]string) template.HTML {
	dataPage := models.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы

	moduleResult := c.ModuleBuild(block, r, dataPage, nil, false)

	return moduleResult.result
}

// Параметры обязательные для задания
// Удаление кешей независимо от контекста текущего процесса (подключаемся к новому неймспейсу)
// &namespace - таблица в reindexer
// &link - связи для выборки (фиксируем uid-страницы и uid-блока) (может быть значение all - удалить все значения кеша)
//func ClearCache(w http.ResponseWriter, r *http.Request) {
//
//	var err error
//	var countDeleted int
//	status := "OK"
//	ns 		:= r.FormValue("namespace")
//	link 	:= r.FormValue("link")
//
//	if ns == "" || link == "" {
//		ResponseJSON(w, "Parametrs: &namespace=, &link=", "ErrorNullParameter", err, nil)
//		return
//	}
//
//	ns = strings.Replace(ns, "/", "_", -1) //заменяем для имен приложений из ru/ru в формат ru_ru
//	if ns == "" {
//		ns = Namespace
//	}
//
//	DBCache_clear := reindexer.NewReindex(BaseCache)
//	err = DBCache_clear.OpenNamespace(ns, reindexer.DefaultNamespaceOptions(), Value{})
//
//	if link == "all" {
//		// паременты не переданы - удаляем все объекты в заданном неймспейсе
//		countDeleted, err = DBCache_clear.Query(ns).
//			Not().WhereString("Uid", reindexer.EQ, "").Delete()
//	} else {
//		// паременты не переданы - удаляем согласно шаблону
//		countDeleted, err = DBCache_clear.Query(ns).
//			Where("Link", reindexer.SET, link).Delete()
//	}
//
//
//	ResponseJSON(w,  countDeleted, status, err, nil)
//}
//
