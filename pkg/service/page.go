package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"sync"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

var (
	re = regexp.MustCompile("(?m)^\\s+")
)

// Page ...
func (s *service) Page(ctx context.Context, in model.ServiceIn) (out model.ServicePageOut, err error) {
	defer s.timingService("Page", time.Now())
	defer s.errorMetric("Page", err)

	var objPages models.ResponseData
	var objPage *models.ResponseData

	//t := time.Now()
	// ПЕРЕДЕЛАТЬ или на кеширование страниц и на доп.проверку
	if in.Page == "" {
		// получаем все страницы текущего приложения
		objPages, err = s.api.LinkGetWithCache(ctx, s.cfg.TplAppPagesPointsrc, s.cfg.DataUid, "in", "")
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

	//t2 := time.Now()

	// запрос объекта страницы
	objPage, err = s.api.ObjGetWithCache(ctx, in.Page)
	if err != nil {
		err = fmt.Errorf("%s (%s)", "Error: Fail GET-request!", err)
		return out, err
	}

	//log.Printf("\n\nполучаем (с кешем) объект страницы) s.api.ObjGetWithCache %fc reqID: %s\n", time.Since(t2).Seconds(), logger.GetRequestIDCtx(ctx))

	if objPage == nil {
		err = fmt.Errorf("%s (%s)", "Error: Fail GET-request! (response is empty)")
		return out, err
	}

	// ФИКС! иногда в разных приложениях называют одинаково страницы.
	// удаляем из объекта objPage значения не текущего приложения
	if len(objPage.Data) > 1 {
		for k, v := range objPage.Data {
			app, _ := v.Attr("app", "src")
			if app != s.cfg.DataUid {
				lib.RemoveElementFromData(objPage, k)
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

	//fmt.Printf("\nсгенерили подготовительные объекты (макет и страница)  values %fc", time.Since(t).Seconds())
	//t3 := time.Now()

	out.Body, err = s.BPage(ctx, in, *objPage, values)

	//fmt.Printf("\nгенерация страницы (сервис) s.BPage %fc\n", time.Since(t3).Seconds())

	return out, err
}

// BPage собираем страницу
func (s *service) BPage(ctx context.Context, in model.ServiceIn, objPage models.ResponseData, values map[string]interface{}) (result string, err error) {
	var objMaket *models.ResponseData
	var objBlocks models.ResponseData
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
	objBlocks, err = s.api.LinkGetWithCache(ctx, s.cfg.TplAppBlocksPointsrc, pageUID, "in", "")
	if err != nil {
		return result, fmt.Errorf("error. blocks is not found for this page: %s, err: %s", pageUID, err)
	}
	//if len(objBlocks.Data) != 0 {
	//	return result, fmt.Errorf("error. blocks is not found for this page: %s, tpl: %s", pageUID, s.cfg.TplAppBlocksPointsrc)
	//}

	// 3 запрос на объект макета
	objMaket, err = s.api.ObjGetWithCache(ctx, maketUID)
	if objMaket == nil {
		return result, fmt.Errorf("%s (uid: %s)", "Error. Object maket is empty.", maketUID)
	}
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

	if s.cfg.BuildModuleParallel.Value && 0 == 1 {
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
			vv.Copies = v.Copies
			vv.Parent = v.Parent
			idBlock, _ := v.Attr("id", "value") // название блока

			// проверяем на налиие прав на показ блока
			publishRoles, _ := v.Attr("publishroles", "src")
			if publishRoles != "" {

				// ПРОВЕРА ПРАВ ПУБЛИКАЦИИ
				// получаем значение профиля из контекста
				profile, ok := ctx.Value("profile").(models.ProfileData)
				if ok {
					if !strings.Contains(publishRoles, profile.CurrentRole.Uid) {
						continue
					}
				}
			}

			if strings.Contains(shemaJSON, idBlock) { // наличие этого блока в схеме
				wg.Add(1)
				go s.block.GetToChannel(ctx, in, vv, page, values, buildChan, &wg)
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
			if strings.Contains(shemaJSON, v.Id) { // наличие этого блока в схеме
				moduleResult, err = s.block.Get(ctx, in, v, page, values)
				if err != nil {
					logger.Error(ctx, fmt.Sprintf("[BPage] Error generate page, title: %s, id: %s", page.Title, page.Id), zap.Error(err))
				}

				p.Blocks[v.Id] = moduleResult.Result
				p.Stat = append(p.Stat, moduleResult.Stat)
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

	var dataFile string
	if maketFileInside != "" {
		dataFile = maketFileInside
	} else {
		if maketFile == "" {
			return "", fmt.Errorf("error. path maketFile is empty")
		}
		sliceMake := strings.Split(maketFile, "/")
		if len(sliceMake) < 4 {
			return "", fmt.Errorf("error path maketFile. current maketFile: %s", maketFile)
		}
		maketFile = strings.Join(sliceMake[3:], "/")

		byteFile, _, err := s.vfs.Read(ctx, maketFile)
		if err != nil {
			logger.Error(ctx, fmt.Sprintf("error vfs.Read, maketFile: %s", maketFile), zap.Error(err))
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
	result = re.ReplaceAllString(result, "")

	return
}
