package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
)

// Page ...
func (s *service) Page(ctx context.Context, in model.ServiceIn) (out model.ServicePageOut, err error) {
	var objPages, objPage models.ResponseData

	// ПЕРЕДЕЛАТЬ или на кеширование страниц и на доп.проверку
	if in.Page == "" {
		// получаем все страницы текущего приложения
		objPages, err = s.api.LinkGet(s.cfg.TplAppPagesPointsrc, s.cfg.DataUid, "in", "")
		//s.tree.Curl("GET", "_link?obj="+s.cfg.DataUid+"&source="+s.cfg.TplAppPagesPointsrc+"&mode=in", "", &objPages, map[string]string{})

		for _, v := range objPages.Data {
			if def, _ := v.Attr("default", "value"); def == "checked" {
				fmt.Println(v.Attr("app", "src"))
				if appUid, _ := v.Attr("app", "src"); appUid == s.cfg.UidService {
					in.Page = v.Uid
				}
			}
		}
	}

	if in.Page == "" {
		ff, _ := json.Marshal(objPages)
		err_url := "_link?obj=" + s.cfg.DataUid + "&source=" + s.cfg.TplAppPagesPointsrc + "&mode=in"
		err = fmt.Errorf("%s", "Error: not default page ("+string(ff)+") (url:"+err_url+", orm: "+s.cfg.UrlApi+")")
		return out, err
	}

	// запрос объекта страницы
	objPage, err = s.api.ObjGet(in.Page)
	//_, err = s.tree.Curl("GET", "_objs/"+in.Page, "", &objPage, map[string]string{})
	if err != nil {
		err = fmt.Errorf("%s (%s)", "Error: Fail GET-request!", err)
		return out, err
	}

	// ФИКС! иногда в разных приложениях называют одинаково страницы.
	// удаляем из объекта objPage значения не текущего приложения
	if len(objPage.Data) > 1 {
		for k, v := range objPage.Data {
			app, _ := v.Attr("app", "src")
			if app != s.cfg.DataUid {
				lib.RemoveElementFromData(&objPage, k)
			}
		}
	}

	// формируем значение переменных, переданных в страницу
	values := map[string]interface{}{}

	values["Prefix"] = s.cfg.ClientPath + s.cfg.PathTemplates
	values["Domain"] = s.cfg.Domain
	values["Path"] = s.cfg.ClientPath
	values["CDN"] = ""
	values["Title"] = s.cfg.Title
	values["URL"] = in.Url
	values["Referer"] = in.Referer
	values["RequestURI"] = in.RequestURI
	values["Profile"] = in.Profile
	values["Cookie"] = in.RequestRaw.Cookies()
	values["Request"] = in.RequestRaw

	out.Body, err = s.BPage(ctx, in, objPage, values)

	return out, err
}

// Собираем страницу
func (s *service) BPage(ctxp context.Context, in model.ServiceIn, objPage models.ResponseData, values map[string]interface{}) (result string, err error) {
	var objMaket, objBlocks models.ResponseData
	var t *template.Template
	moduleResult := model.ModuleResult{}
	//statModule := map[string]interface{}{}

	// флаг режима генерации модулей (последовательно/параллельно)
	p := &model.Page{}
	p.Title = s.cfg.Title
	p.Domain = s.cfg.Domain
	p.Metric = template.HTML(s.cfg.Metric)
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
		return "", fmt.Errorf("%s", "Error: Object page is null.")
	}

	pageUID := objPage.Data[0].Uid
	maketUID, _ := objPage.Data[0].Attr("maket", "src")
	page := objPage.Data[0]

	// 1.0 проверка на принадлежность страницы текущему проекту
	// ДОДЕЛАТЬ СРОЧНО!!!

	// 2 запрос на объекты блоков страницы
	objBlocks, err = s.api.LinkGet(s.cfg.TplAppBlocksPointsrc, pageUID, "in", "")

	// 3 запрос на объект макета
	objMaket, err = s.api.ObjGet(maketUID)
	//s.tree.Curl("GET", "_objs/"+maketUID, "", &objMaket, map[string]string{})

	if len(objMaket.Data) == 0 {
		return result, fmt.Errorf("%s (uid: %s)", "Error. Object maket is empty.", maketUID)
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
		return "", fmt.Errorf("%s (%s)", "Error! Fail json shema!", err)
	}
	p.Shema = i

	// 4 запускаем сборку модулей (получаем сгенерированный template.HTML без JS и CSS
	// шаблоны рендерятся в каждом модуле отдельно (можно далее хранить в кеше)

	if s.cfg.BuildModuleParallel.Value && 1 == 1 {
		ctx := context.WithValue(context.Background(), "timeout", s.cfg.TimeoutBlockGenerate.Value)
		ctx, cancel := context.WithCancel(ctx)

		// ПАРАЛЛЕЛЬНО
		wg := sync.WaitGroup{}
		var buildChan = make(chan model.ModuleResult, len(objBlocks.Data))

		for _, v := range objBlocks.Data {
			var vv = models.Data{}
			vv.Id = v.Id
			vv.Uid = v.Uid
			vv.Attributes = v.Attributes
			vv.Title = v.Title
			vv.Source = v.Source
			vv.Rev = v.Rev
			vv.Type = v.Type
			vv.Сopies = v.Сopies
			vv.Parent = v.Parent
			idBlock, _ := v.Attr("id", "value") // название блока

			// проверяем на налиие прав на показ блока
			publishRoles, _ := v.Attr("publishroles", "src")
			if publishRoles != "" {

				// ПРОВЕРА ПРАВ ПУБЛИКАЦИИ
				// получаем значение профиля из контекста
				profile, ok := ctxp.Value("profile").(models.ProfileData)
				if ok {
					if !strings.Contains(publishRoles, profile.CurrentRole.Uid) {
						continue
					}
				}
			}

			if strings.Contains(shemaJSON, idBlock) { // наличие этого блока в схеме
				wg.Add(1)
				go s.GetBlockToChannel(ctx, in, vv, page, shemaJSON, values, buildChan, &wg)
			}
		}

		// ждем завершения интервала и вызываем завершение контекста для запущенных воркеров
		exitTimer := make(chan struct{})
		timerBlockGen := time.NewTimer(s.cfg.TimeoutBlockGenerate.Value)

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
			p.Blocks[k.Id] = k.Result
			p.Stat = append(p.Stat, k.Stat)
		}
	} else {
		// ПОСЛЕДОВАТЕЛЬНО
		for _, v := range objBlocks.Data {
			moduleResult, err = s.GetBlock(in, v, page, shemaJSON, values)
			if err != nil {
				s.logger.Error(err, "[BPage] Error generate page ", page.Title+"("+page.Id+")")
			}

			p.Blocks[v.Id] = moduleResult.Result
			p.Stat = append(p.Stat, moduleResult.Stat)
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

	var dataFile string
	if maketFileInside != "" {
		dataFile = maketFileInside
	} else {
		sliceMake := strings.Split(maketFile, "/")
		maketFile = strings.Join(sliceMake[3:], "/")

		byteFile, _, err := s.vfs.Read(maketFile)
		if err != nil {
			s.logger.Error(err, "error vfs.Read, maketFile", maketFile)
		}
		dataFile = string(byteFile)
	}

	tmp := template.New(maketFile)
	t, err = tmp.Parse(string(dataFile))

	//maketFile = s.cfg.Workingdir + "/" + maketFile
	//pt, err := template.ParseFiles(maketFile)
	//if err != nil {
	//	s.logger.Error(err, "Error ParseFiles (", maketFile, ")")
	//	//fmt.Println(err, maketFile)
	//}

	// в режиме отладки пересборка шаблонов происходит при каждом запросе
	//if !s.cfg.CompileTemplates.Value {
	//t = template.Must(template.New(maketFile).Funcs(funcMap).ParseFiles(maketFile))
	//t = template.Must(pt, err)
	t.Execute(&c, p)
	//} else {
	//t.ExecuteTemplate(&c, maketFile, p)
	//}

	result = c.String()

	// чистим от лишних пробелов
	re := regexp.MustCompile("(?m)^\\s+")
	result = re.ReplaceAllString(result, "")

	return
}

// получаем содержимое блока в передачей через канал
func (s *service) GetBlockToChannel(ctx context.Context, in model.ServiceIn, block, page models.Data, shemaJSON string, values map[string]interface{}, buildChan chan model.ModuleResult, wg *sync.WaitGroup) (err error) {
	defer wg.Done()

	// проверка на выход по сигналу
	select {
	case <-ctx.Done():
		return
	default:
	}

	moduleResult, err := s.GetBlock(in, block, page, shemaJSON, values)
	if err != nil {
		moduleResult.Err = err
		moduleResult.Result = template.HTML(fmt.Sprint(err))
	}
	buildChan <- moduleResult

	return
}

// получение содержимого блока (с учетом операций с кешем)
func (s *service) GetBlock(in model.ServiceIn, block, page models.Data, shemaJSON string, values map[string]interface{}) (moduleResult model.ModuleResult, err error) {
	var addСonditionPath bool
	var addСonditionURL bool

	cacheInt, _ := block.Attr("cache", "value") // включен ли режим кеширования
	cache_nokey2, _ := block.Attr("cache_keyAddPath", "value")
	cache_nokey3, _ := block.Attr("cache_keyAddURL", "value")

	if cache_nokey2 == "checked" {
		addСonditionPath = true
	}
	if cache_nokey3 == "checked" {
		addСonditionURL = true
	}

	//t1 := time.Now()

	if strings.Contains(shemaJSON, block.Id) { // наличие этого блока в схеме

		// если интервал не задан, то не кешируем
		cacheInterval, err := strconv.Atoi(cacheInt)
		if err != nil {
			cacheInterval = 0
		}

		// если включен кеш и есть интервал кеширования
		if s.cache.Active() && cacheInterval != 0 {

			// читаем из кеша и отдаем (ВСЕГДА сразу)
			key, cacheParams := s.cache.GenKey(block.Uid, in.CachePath, in.CacheQuery, addСonditionPath, addСonditionURL)
			result, _, flagExpired, err := s.cache.Read(key)

			//fmt.Println("read:", time.Since(t1), block.Id, key)

			// 1 кеша нет (срабатывает только при первом формировании)
			if err != nil {
				//fmt.Println("genr NULL:", time.Since(t1), block.Id, key, err, result)

				result, err = s.updateCache(key, cacheParams, cacheInterval, in, block, page, values)
			} else {
				// 2 время закончилось (не обращаем внимание на статус "обновляется" потому, что при изменении статуса на "обновляем"
				// мы увеличиваем время на предельно время проведения обновления
				// требуется обновить фоном (отдали текущие данные из кеша)
				if flagExpired {
					//fmt.Println("genr flagExpired:", time.Since(t1), block.Id, key, flagExpired)

					go s.updateCache(key, cacheParams, cacheInterval, in, block, page, values)
				}
			}

			moduleResult = model.ModuleResult{
				Id:     block.Id,
				Result: template.HTML(result),
				Stat:   nil,
				Err:    nil,
			}

		} else {
			mResult, err := s.block.Generate(in, block, page, values)
			if err != nil {
				moduleResult.Result = ""
				moduleResult.Err = err
				return moduleResult, err
			}

			moduleResult = mResult
		}

	} else {
		s.logger.Error(nil, "Error. Block"+block.Id+" from page "+page.Id+" in not used.")
		//fmt.Println("fail: ", block.Id)
	}

	//fmt.Println("Time:", time.Since(t1), "Cache:", s.cache.Active(), "Block:", block.Id)

	return
}

// внутренняя фунция сервиса.
// не вынесена в пакет Cache потому-что требуется генерировать блок
func (s *service) updateCache(key, cacheParams string, cacheInterval int, in model.ServiceIn, block models.Data, page models.Data, values map[string]interface{}) (result string, err error) {
	t1 := time.Now()

	err = s.cache.SetStatus(key, "updated")
	if err != nil {
		result = fmt.Sprint(err)
		fmt.Println("err ", block.Id, err)
	}

	moduleResult, err := s.block.Generate(in, block, page, values)
	if err != nil {
		result = fmt.Sprintf("Error [Generate] in updateCache from %s. Cache not saved. Time generate: %s. Error: %s", block.Id, time.Since(t1), err)
		fmt.Println(result)
		return result, err
	}

	err = s.cache.Write(key, cacheParams, cacheInterval, block.Uid, page.Uid, string(moduleResult.Result))
	if err != nil {
		result = fmt.Sprint(err)
		fmt.Println("err ", block.Id, time.Since(t1), err)
	}

	//fmt.Println("save:", time.Since(t1), block.Id, key, err)

	result = string(moduleResult.Result)

	return
}

// возвращаем сформированную страницу в template.HTML (для cockpit-a и dashboard)
//func (s *service) TIndex(w http.ResponseWriter, r *http.Request, Config map[string]string) template.HTML {
//
//	var objPage, objApp models.ResponseData
//	vars := mux.Vars(r)
//	page := vars["obj"] // ид-страницы передается через переменную obj
//
//	// указатель на профиль текущего пользователя
//	ctx := r.Context()
//	var profile model.ProfileData
//	profileRaw := ctx.Value("UserRaw")
//	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)
//
//
//	// можем задать также через &page=страница
//	if r.FormValue("page") != "" {
//		page = r.FormValue("page")
//	}
//
//	if page == "" {
//		return ""
//	}
//
//	// заменяем значения при вызове ф-ции из GUI ибо они пустые, ведь приложение полностью не инициализировано через конфиг
//
//	if page == "" {
//		return template.HTML("Error: Not id page")
//	}
//
//	// запрос объекта страницы
//	s.tree.Curl("GET", "_objs/"+page, "", &objPage)
//
//	//fmt.Println("objPage: ", objPage)
//
//	if &objPage == nil {
//		return template.HTML("Error: Not found page-object.") // если не найден объект страницы
//	}
//
//	if len(objPage.Data) == 0 {
//		return template.HTML("Error: Not found page-object.") // если не найден объект страницы
//	}
//
//	// Uid-приложения
//	appUid, found := objPage.Data[0].Attr("app", "src")
//	if !found {
//		return template.HTML("Error: Not selected application from this page.")
//	}
//
//	// запрос объекта приложения
//	s.tree.Curl("GET", "_objs/"+appUid, "", &objApp)
//	if &objApp == nil {
//		return template.HTML("Error: Not found application-object.") // если не найден объект приложения
//	}
//
//	//fmt.Println("objApp: ", objApp)
//
//	// получаем значения аттрибутов для данного приложения
//	path_template, found := objApp.Data[0].Attr("path_templates", "value")
//	if !found {
//		return template.HTML("Error: Not selected path_templates from this application.")
//	}
//
//	// получаем значения аттрибутов для данного приложения
//	tpl_app_blocks_pointsrc, found := objApp.Data[0].Attr("tpl_app_blocks", "src")
//	if !found {
//		return template.HTML("Error: Not selected tpl_app_blocks from this application.")
//	}
//
//	//pp := strings.Split(Domain, "/")
//	//if len(pp) == 1 {
//	//	ClientPath = Domain + "/" + "gui"
//	//}
//
//	// получили значение Request в json - возвращаем в http.Request
//	//var PageRequest *http.Request
//	//json.Unmarshal([]byte(JRequest), &PageRequest)
//
//	// формируем значение переменных, переданных в страницу
//	values := map[string]interface{}{}
//	values["Prefix"] = s.cfg.ClientPath + path_template
//	values["Domain"] = s.cfg.Domain
//	values["Path"] = s.cfg.ClientPath
//	values["CDN"] = ""
//	values["Title"] = s.cfg.Title
//	values["URL"] = r.URL.Query().Encode()
//	values["Referer"] = r.Referer()
//	values["RequestURI"] = r.RequestURI
//	values["Profile"] = profile
//
//
//	result := s.BPage(in, tpl_app_blocks_pointsrc, objPage, values)
//
//	return template.HTML(result)
//}

// генерируем один блок через внутренний запрос - для cocpit-a
//func (s *service) TBlock(r *http.Request, block model.Data, Config map[string]string) template.HTML {
//	dataPage 		:= model.Data{} // пустое значение, используется в блоке для кеширования если он вызывается из страницы
//	moduleResult := s.ModuleBuild(block, r, dataPage, nil, false)
//
//	return moduleResult.Result
//}

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
