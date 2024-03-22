package block

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

func (b *block) generate(ctx context.Context, in model.ServiceIn, block models.Data, page models.Data, values map[string]interface{}) (result model.ModuleResult, err error) {
	var c bytes.Buffer
	result.Id = block.Id

	// обработка всех странных ошибок
	// ВКЛЮЧИТЬ ПОЗЖЕ!!!!
	defer func() {
		if er := recover(); er != nil {
			//ft, err := template.ParseFiles("./upload/control/templates/errors/503.html")
			//if err != nil {
			//	l.Logger.Error(err)
			//}
			//t = template.Must(ft, nil)

			result.Result = template.HTML(fmt.Sprint(er))
			result.Err = fmt.Errorf("%s", er)
			return
		}
	}()

	t1 := time.Now()

	// заменяем в State localhost на адрес домена (если это подпроцесс то все норм, но если это корневой сервис,
	// то у него url_proxy - localhost и узнать реньше адрес мы не можем, ибо еще домен не инициировался
	// а значит подменяем localhost при каждом обращении к модулю
	if strings.Contains(b.cfg.ProxyPointsrc, "localhost") {
		b.cfg.ProxyPointsrc = "//" + in.Host
	}

	bl := model.Block{}
	bl.Mx.Lock()
	defer bl.Mx.Unlock()

	stat := map[string]interface{}{}
	stat["start"] = t1
	stat["status"] = "OK"
	stat["title"] = block.Title
	stat["id"] = block.Id

	bl.Value = map[string]interface{}{}

	dataSet := make(map[string]interface{})
	dataname := "default" // значение по-умолчанию (будет заменено в потоках)

	tplName, _ := block.Attr("module", "src")
	tquery, _ := block.Attr("tquery", "src")

	// //////////////////////////////////////////////////////////////////////////////
	// в блоке есть настройки поля расширенного фильтра, который можно добавить в самом блоке
	// дополняем параметры request-a, доп. параметрами, которые переданы через блок
	extfilter, _ := block.Attr("extfilter", "value") // дополнительный фильтр для блока
	dv := []models.Data{block}
	extfilter, err = b.function.Exec(extfilter, dv, bl.Value, in, block.Id+"_extfilter")
	if err != nil {
		logger.Error(ctx, "[Generate] Error parsing Exec from block.")
		fmt.Println("[Generate] Error parsing Exec from block: ", err)
	}

	extfilter = strings.Replace(extfilter, "?", "", -1)

	// парсим переденную строку фильтра
	m, err := url.ParseQuery(extfilter)
	if err != nil {
		logger.Error(ctx, "[Generate] Error parsing extfilter from block.", zap.Error(err))
		fmt.Println("[Generate] Error parsing extfilter from block.: ", err)
	}

	// добавляем в URL переданное значение из настроек модуля
	// если этих значений еще нет (НЕ ЗАМЕНЯЕМ)
	//var q url.Values
	var blockQuery = in.Query // Get a copy of the query values.
	for k, v := range m {
		if _, found := blockQuery[k]; !found {
			blockQuery.Add(k, strings.Join(v, ",")) // Add a new value to the set. Переводим обратно в строку из массива
		}
	}
	// //////////////////////////////////////////////////////////////////////////////

	tconfiguration, _ := block.Attr("configuration", "value")
	tconfiguration = strings.Replace(tconfiguration, "  ", "", -1)

	uuid := lib.UUID()

	if values != nil && len(values) != 0 {
		for k, v := range values {
			if _, found := bl.Value[k]; !found {
				bl.Value[k] = v
			}
		}
	}

	bl.Value["Rand"] = uuid[1:6] // переопределяем отдельно для каждого модуля
	bl.Value["URL"] = in.Url
	bl.Value["Prefix"] = "/" + b.cfg.Domain + b.cfg.PathTemplates
	bl.Value["Domain"] = b.cfg.Domain
	bl.Value["CDN"] = b.cfg.UrlFs
	bl.Value["Path"] = b.cfg.ClientPath
	bl.Value["Title"] = b.cfg.Title
	bl.Value["Form"] = in.Form
	bl.Value["RequestURI"] = in.RequestURI
	bl.Value["Referer"] = in.Referer
	bl.Value["Profile"] = in.Profile
	bl.Value["Cookie"] = in.RequestRaw.Cookies()
	bl.Value["Request"] = in.RequestRaw

	//fmt.Println("tconfiguration: block", block.Id, tconfiguration, "\n")

	// обработк @-функции в конфигурации
	dv = []models.Data{block}
	dogParseConfiguration, err := b.function.Exec(tconfiguration, dv, bl.Value, in, block.Id)
	if err != nil {
		mes := "[Generate] Error DogParse configuration: (" + fmt.Sprint(err) + ") " + tconfiguration
		result.Result = b.moduleError(ctx, mes)
		result.Err = err
		logger.Error(ctx, mes, zap.Error(err))

		return
	}

	//fmt.Println("block", block.Id, tconfiguration, "\ndogParseConfiguration: ", dogParseConfiguration, "\n\n\n")

	// конфигурация без обработки @-функции
	var confRaw map[string]model.Element
	if tconfiguration != "" {
		err = json.Unmarshal([]byte(tconfiguration), &confRaw)
	}
	if err != nil {
		mes := "[Generate] Error Unmarshal configuration: (" + fmt.Sprint(err) + ") " + tconfiguration
		result.Result = b.moduleError(ctx, "[Generate] Error Unmarshal configuration: ("+fmt.Sprint(err)+") "+tconfiguration)
		result.Err = err
		logger.Error(ctx, mes, zap.Error(err))

		return
	}

	// конфигурация с обработкой @-функции
	var conf map[string]model.Element
	if dogParseConfiguration != "" {
		err = json.Unmarshal([]byte(dogParseConfiguration), &conf)
	}
	if err != nil {
		mes := "[Generate] Error json-format configurations: (" + fmt.Sprint(err) + ") " + dogParseConfiguration
		result.Result = b.moduleError(ctx, "[Generate] Error json-format configurations: ("+fmt.Sprint(err)+") "+dogParseConfiguration)
		result.Err = err
		logger.Error(ctx, mes, zap.Error(err))

		return
	}

	// сформировал структуру полученных описаний датасетов
	var source []map[string]string
	if d, found := conf["datasets"]; found {
		rm, _ := json.Marshal(d.Source)
		err = json.Unmarshal(rm, &source)

		if err != nil {
			stat["status"] = "error"
			stat["description"] = fmt.Sprint(err)

			result.Result = b.moduleError(ctx, err)
			result.Err = err
			result.Stat = stat
			mes := "[Generate] Error generate datasets."
			logger.Error(ctx, mes, zap.Error(err))

			return result, err
		}
	}

	// ПЕРЕДЕЛАТЬ НА ПАРАЛЛЕЛЬНЫЕ ПОТОКИ
	if tquery != "" {
		slquery := strings.Split(tquery, ",")

		var name, uid string
		for _, queryUID := range slquery {

			// подставляем название датасета из конфигурации
			for _, v1 := range source {
				if _, found := v1["name"]; found {
					name = v1["name"]
				}
				if _, found := v1["uid"]; found {
					uid = v1["uid"]
				}
				if uid == queryUID {
					dataname = name
				}
			}

			//fmt.Println(queryUID, dataname, source, in.Token, blockQuery.Encode(), in.Method, in.PostForm)
			ress := b.queryWorker(ctx, queryUID, dataname, source, in.Token, blockQuery.Encode(), in.Method, in.PostForm) //in.QueryRaw
			//fmt.Println(ress)

			dataSet[dataname] = ress
		}
	}

	logger.Info(ctx, "gen block",
		zap.String("step", "подготовка к генерации"),
		zap.Float64("timing", time.Since(t1).Seconds()),
		zap.String("block", block.Id), zap.String("rnd", uuid))

	bl.Data = dataSet
	bl.Page = page
	bl.Configuration = conf
	// b.ConfigurationRaw = confRaw
	bl.ConfigurationRaw = tconfiguration
	//bl.Request = r

	result.Id = block.Id

	// если содержится разделитель - значит передан путь к файлу (старая версия) и генерируем из файла
	// иначе берем значение из поля codetpl (новая версия), если пусто, то из поля _filecontent_url
	// (для случаем, когда блок выбрали, но содержимое файла не перенесли в новое поле и оно хрантся в поле автосохранения файла)
	if strings.Contains(tplName, sep) {
		//tplName = b.cfg.Workingdir + "/" + tplName

		c, err = b.generateBlockFromFile(ctx, tplName, bl)
		if err != nil {
			err = fmt.Errorf("%s file:'%s' (%s)", "Error: Generate Module from file is failed!", tplName, err)
			result.Result = template.HTML(fmt.Sprint(err))
			return result, nil
		}
	} else {
		uidModule, _ := block.Attr("module", "src")
		var objModule *models.ResponseData

		t0 := time.Now()

		// запрос на объект HTML
		objModule, err = b.api.ObjGetWithCache(ctx, uidModule)
		//_, err = b.tree.Curl("GET", "_objs/"+uidModule, "", &objModule, map[string]string{})

		logger.Info(ctx, "gen block", zap.Float64("timing", time.Since(t0).Seconds()),
			zap.String("step", "запрос на объект HTML ObjGetWithCache"),
			zap.String("block", block.Id),
			zap.String("rnd", uuid))

		if err != nil {
			err = fmt.Errorf("%s (%s)", "Error: Get object Module is failed!", err)
			result.Result = template.HTML(fmt.Sprint(err))
			return result, err
		}
		if len(objModule.Data) == 0 {
			err = fmt.Errorf("%s", "Error: Object Module is null!")
			result.Result = template.HTML(fmt.Sprint(err))
			return result, err
		}

		// если выбрано несколько блоков, их все объединяем в один (очередность случайная)
		htmlCode := ""
		for _, v := range objModule.Data {
			codetpl, _ := v.Attr("codetpl", "value")
			if codetpl == "" {
				codetpl, _ = v.Attr("_filecontent_module", "value")
				if codetpl == "" {
					codetpl, _ = v.Attr("_filecontent_url", "value")
				}
			}
			htmlCode = htmlCode + codetpl
		}

		t2 := time.Now()
		c, err = b.generateBlockFromField(htmlCode, bl, block.Id)

		logger.Info(ctx, "gen block", zap.Float64("timing", time.Since(t2).Seconds()),
			zap.String("step", "generateBlockFromField full"),
			zap.String("block", block.Id),
			zap.String("rnd", uuid))

	}

	// ошибка при генерации страницы
	if err != nil {
		mes := fmt.Sprintf("Error. Generate module is failed. err: (%s)", err)
		logger.Error(ctx, mes, zap.Error(err))
		result.Result = template.HTML(mes)
		result.Id = block.Id

		return result, nil
	}

	blockBody := c.String()

	// чистим от лишних пробелов
	re := regexp.MustCompile("(?m)^\\s+")
	blockBody = re.ReplaceAllString(blockBody, "")

	result.Result = template.HTML(blockBody)
	result.Stat = stat

	logger.Info(ctx, "gen block", zap.Float64("timing", time.Since(t1).Seconds()),
		zap.String("step", "full"),
		zap.String("block", block.Id),
		zap.String("rnd", uuid))

	return result, err
}

// generateBlockFromFile генерируем блок из файла (для совместимости со старыми модулями)
func (b *block) generateBlockFromFile(ctx context.Context, tplName string, bl model.Block) (c bytes.Buffer, err error) {
	var tmpl *template.Template

	dataFile, _, err := b.vfs.Read(ctx, b.clearPath(tplName))
	if err != nil {
		err = fmt.Errorf("%s", "error read file from vfs. path: %s", b.clearPath(tplName))
		return
	}

	tmpl = template.New(tplName).Funcs(b.tplfunc.GetFuncMap())
	_, err = tmpl.Parse(string(dataFile))
	if err != nil {
		err = fmt.Errorf("%s", "Error: Getting path.Base failed! tplName: %s", tplName)
		return
	}

	if &bl != nil && &c != nil {
		if tmpl == nil {
			err = fmt.Errorf("%s", "Error: Parsing template file is fail!")
		} else {
			err = tmpl.Execute(&c, bl)
		}
	} else {
		err = fmt.Errorf("%s", "Error: Generate data block is fail!")
	}

	return c, err
}

// GenerateBlockFromField генерируем блок из переданного текста
func (b *block) generateBlockFromField(value string, bl model.Block, block string) (c bytes.Buffer, err error) {
	t1 := time.Now()
	rnd := lib.UUID()
	tmpl, err := template.New("name").Funcs(b.tplfunc.GetFuncMap()).Parse(value)
	if err != nil {
		return
	}

	logger.Info(context.Background(), "gen block",
		zap.String("step", "Parse"),
		zap.String("block", block),
		zap.Float64("timing", time.Since(t1).Seconds()), zap.String("rnd", rnd))

	t2 := time.Now()

	err = tmpl.Execute(&c, bl)

	logger.Info(context.Background(), "gen block", zap.String("step", "Execute"),
		zap.String("block", block),
		zap.Float64("timing", time.Since(t2).Seconds()), zap.String("rnd", rnd))

	return c, err
}

// ErrorModuleBuild вываливаем ошибку при генерации модуля
func (b *block) errorModuleBuild(ctx context.Context, stat map[string]interface{}, buildChan chan model.ModuleResult, timerRun interface{}, errT error) {
	var result model.ModuleResult

	stat["cache"] = "false"
	stat["time"] = timerRun
	result.Stat = stat
	result.Result = template.HTML(fmt.Sprint(errT))
	result.Err = errT

	buildChan <- result

	return
}

// queryUID - ид-запроса
func (b *block) queryWorker(ctx context.Context, queryUID, dataname string, source []map[string]string, token, queryRaw, metod string, postForm url.Values) interface{} {
	//var resp Response

	resp := b.guiQuery(ctx, queryUID, token, queryRaw, metod, postForm)

	//switch x := resp1.(type) {
	//case Response:
	//	resp = resp1.(Response)
	//
	//default:
	//	resp.Data = resp1ч
	//}

	///////////////////////////////////////////
	// Расчет пагенации
	///////////////////////////////////////////

	var m3 models.Response
	b1, _ := json.Marshal(resp)
	json.Unmarshal(b1, &m3)
	var last, current, from, to, size int
	var list []int

	resultLimit := m3.Metrics.ResultLimit
	resultOffset := m3.Metrics.ResultOffset
	size = m3.Metrics.ResultSize

	if size != 0 && resultLimit != 0 {
		j := 0
		for i := 0; i <= size; i = i + resultLimit {
			j++
			list = append(list, j)
			if i >= resultOffset && i < resultOffset+resultLimit {
				current = j
			}
		}
		last = j
	}

	from = current*resultLimit - resultLimit
	to = from + resultLimit

	// подрезаем список страниц
	lFrom := 0
	if current != 1 {
		lFrom = current - 2
	}
	if lFrom <= 0 {
		lFrom = 0
	}

	lTo := current + 4
	if lTo > last {
		lTo = last
	}
	if lTo <= 0 {
		lTo = 0
	}

	lList := list[lFrom:lTo]

	resp.Metrics = m3.Metrics
	resp.Metrics.PageLast = last
	resp.Metrics.PageCurrent = current
	resp.Metrics.PageList = lList

	resp.Metrics.PageFrom = from
	resp.Metrics.PageTo = to

	///////////////////////////////////////////
	///////////////////////////////////////////

	return resp
}

// ErrorPage вывод ошибки выполнения блока
func (b *block) errorPage(ctx context.Context, err interface{}, w http.ResponseWriter, r *http.Request) {
	p := model.ErrorForm{
		Err: err,
		R:   *r,
	}
	logger.Info(ctx, "get ErrorPage", zap.Error(fmt.Errorf("err: %s", err)))

	t := template.Must(template.ParseFiles("./upload/control/templates/errors/500.html"))
	err = t.Execute(w, p)
	if err != nil {
		logger.Error(ctx, "error Execute in ErrorPage", zap.Error(fmt.Errorf("err: %s", err)))
	}
}

// ModuleError вывод ошибки выполнения блока
func (l *block) moduleError(ctx context.Context, err interface{}) template.HTML {
	var c bytes.Buffer

	p := model.ErrorForm{
		Err: err,
	}

	logger.Info(ctx, "ModuleError", zap.Error(fmt.Errorf("err: %s", err)))
	//fmt.Println("ModuleError: ", err)

	wd := l.cfg.Workingdir
	t := template.Must(template.ParseFiles(wd + "/upload/control/templates/errors/503.html"))

	t.Execute(&c, p)
	result := template.HTML(c.String())

	return result
}

// GUIQuery отправка запроса на получения данных из интерфейса GUI
// параметры переданные в строке (r.URL) отправляем в теле запроса
func (b *block) guiQuery(ctx context.Context, tquery, token, queryRaw, method string, postForm url.Values) (returnResp models.Response) {
	var err error
	bodyJSON, _ := json.Marshal(postForm)

	// добавляем к пути в запросе переданные в блок параметры ULR-а (возможно там есть параметры для фильтров)
	filters := queryRaw
	if filters != "" {
		filters = "?" + filters
	}

	// ФИКС!
	// добавляем еще токен (cookie) текущего пользователя
	// это нужно для случая, если мы вызываем запрос из запроса и кука не передается
	// а если куки нет, то сбрасывается авторизация
	if token != "" {
		if strings.Contains(filters, "?") {
			filters = filters + "&iam=" + token
		} else {
			filters = filters + "?iam=" + token
		}
	}

	resultInterface, _ := b.api.Query(ctx, tquery+filters, method, string(bodyJSON))

	// попытка в ResponseData
	var dd1 models.ResponseData
	err = json.Unmarshal([]byte(resultInterface), &dd1)

	// нам тут нужен Response, но бывают внешние запросы,
	// поэтому если не Response то дописываем в Data полученное тело
	if err == nil {
		returnResp.Data = dd1.Data
		returnResp.Metrics = dd1.Metrics
		returnResp.Status = dd1.Status

		return returnResp
	}

	// попытка в Response
	var dd2 models.Response
	err = json.Unmarshal([]byte(resultInterface), &dd2)
	if err == nil {
		returnResp.Data = dd1.Data
		returnResp.Metrics = dd1.Metrics
		returnResp.Status = dd1.Status

		return returnResp
	}

	// иначе просто передаем в Data данные поученные из неформатного запроса
	returnResp.Data = resultInterface

	return returnResp
}

// clearPath чистим от части пути, от корня домена
func (l *block) clearPath(file string) string {
	// для приложения домен (проект) отличается от пути
	// поэтому удаляем первые две части пути - это точно принадлежит домену
	fileSlice := strings.Split(file, sep)
	if len(fileSlice) > 3 {
		file = strings.Join(fileSlice[3:], sep)
	}

	// TODO костыли, удалить после переноса всех проектов на новые пути
	// для совместимости со старым форматом хранения
	// если есть upload - значит будет /upload/buildbox - это тоже удаляем
	if strings.Contains(file, "upload") {
		fileSlice = strings.Split(file, sep)
		if len(fileSlice) > 2 {
			file = strings.Join(fileSlice[2:], sep)
		}
	}

	file = strings.Replace(file, sep+sep, sep, -1)

	return file
}
