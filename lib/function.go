package app_lib

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const sep = string(os.PathSeparator)

func (c *app) hash(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	return sha1_hash
}

func (c *app) UrlAPI() string {
	return c.urlAPI
}

func (c *app) UrlORM() string {
	return c.urlORM
}

// Создаем файл по указанному пути если его нет
func (c *app) CreateFile(path string) {

	// detect if file exists
	var _, err = os.Stat(path)
	var file *os.File

	// delete old file if exists
	if !os.IsNotExist(err) {
		os.RemoveAll(path)
	}

	// create file
	file, err = os.Create(path)
	if c.isError(err) {
		return
	}
	defer file.Close()

	//log.Warning("==> done creating file", path)
}

// функция печати в лог ошибок (вспомогательная)
func (c *app) isError(err error) bool {
	if err != nil {
		// c.Logger.Warning(err.Error())
	}
	return (err != nil)
}

// пишем в файл по указанному пути
func (c *app) WriteFile(path string, data []byte) {

	// detect if file exists and create
	c.CreateFile(path)

	// open file using READ & WRITE permission
	var file, err = os.OpenFile(path, os.O_RDWR, 0644)

	if c.isError(err) {
		return
	}
	defer file.Close()

	// write into file
	_, err = file.Write(data)
	if c.isError(err) {
		return
	}

	// save changes
	err = file.Sync()
	if c.isError(err) {
		return
	}

	//log.Warning("==> done writing to file")
}

func (c *app) Curl(method, urlc, bodyJSON string, response interface{}, cookies []*http.Cookie) (result interface{}, err error) {

	var mapValues map[string]string
	var req *http.Request
	var flagExtRequest bool
	client := &http.Client{}

	if urlc[:4] == "http" {
		flagExtRequest = true
	}

	// дополняем путем до API если не передан вызов внешнего запроса через http://
	if !flagExtRequest {
		// дополняем путем до API если не передан вызов внешнего запроса через http://
		if urlc[:1] != "/" {
			urlc = c.urlORM + "/" + urlc
		} else {
			urlc = c.urlAPI + "/" + urlc
		}
	}

	if method == "" {
		method = "POST"
	}

	method = strings.Trim(method, " ")
	values := url.Values{}
	actionType := ""

	//// c.Logger.Warning("urlc " , urlc)
	//fmt.Println("urlc1 " , urlc)

	// если в гете мы передали еще и json (его добавляем в строку запроса)
	// только если в запросе не указаны передаваемые параметры
	clearUrl := strings.Contains(urlc, "?")

	bodyJSON = strings.Replace(bodyJSON, "  ", "", -1)
	err = json.Unmarshal([]byte(bodyJSON), &mapValues)

	if method == "JSONTOGET" && bodyJSON != "" && clearUrl {
		actionType = "JSONTOGET"
	}
	if method == "JSONTOPOST" && bodyJSON != "" {
		actionType = "JSONTOPOST"
	}

	switch actionType {
	case "JSONTOGET": // преобразуем параметры в json в строку запроса
		if err == nil {
			for k, v := range mapValues {
				values.Set(k, v)
			}
			uri, _ := url.Parse(urlc)
			uri.RawQuery = values.Encode()
			urlc = uri.String()
			req, err = http.NewRequest("GET", urlc, strings.NewReader(bodyJSON))
		} else {
			// c.Logger.Warning("Error! Fail parsed bodyJSON from GET Curl: ", err)
		}
	case "JSONTOPOST": // преобразуем параметры в json в тело запроса

		if err == nil {
			for k, v := range mapValues {
				values.Set(k, v)
			}
			req, err = http.NewRequest("POST", urlc, strings.NewReader(values.Encode()))
			req.PostForm = values
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		} else {
			// c.Logger.Warning("Error! Fail parsed bodyJSON to POST: ", err)
		}
	default:
		req, err = http.NewRequest(method, urlc, strings.NewReader(bodyJSON))
	}

	//req.Header.Add("If-None-Match", `W/"wyzzy"`)
	if err != nil {
		return "", err
	}

	// дополянем куками назначенными для данного запроса
	if cookies != nil {
		for _, v := range cookies {
			req.AddCookie(v)
		}
	}

	resp, err := client.Do(req)
	//fmt.Println(resp.Body, " = ", err)

	if err != nil {
		// c.Logger.Warning("Error request: metod:", method, ", url:", urlc, ", bodyJSON:", bodyJSON)
		return "", err
	} else {
		defer resp.Body.Close()
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	responseString := string(responseData)

	//// c.Logger.Warning("Сделан: ", method, " на: ", urlc, " ответ: ",responseString)

	// возвращаем объект ответа, если передано - в какой объект класть результат
	if response != nil {
		json.Unmarshal([]byte(responseString), &response)
	}

	// всегда отдаем в интерфейсе результат (полезно, когда внешние запросы или сериализация на клиенте)
	json.Unmarshal([]byte(responseString), &result)

	return result, err
}

// ДЛЯ ПОСЛЕДОВАТЕЛЬНОЙ сборки блока
// получаем объект модуля (отображения)
// p 	- объект переданных в модуль данных блока (запрос/конфигураци)
// r 	- значения реквеста
// page - объект страницы, которую парсим
func (l *app) ModuleBuild(block Data, r *http.Request, page Data, values map[string]interface{}, enableCache bool) (result ModuleResult) {
	var err error

	// указатель на профиль текущего пользователя
	ctx := r.Context()
	var profile ProfileData
	profileRaw := ctx.Value("UserRaw")
	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)

	// заменяем в State localhost на адрес домена (если это подпроцесс то все норм, но если это корневой сервис,
	// то у него url_proxy - localhost и узнать реньше адрес мы не можем, ибо еще домен не инициировался
	// а значит подменяем localhost при каждом обращении к модулю
	if strings.Contains(l.ConfigGet("url_proxy"), "localhost") {
		//url_shema := "http"
		//if r.TLS != nil {
		//	url_shema = "https"
		//}
		//l.State["url_proxy"] = url_shema + "://" + r.Host
		l.ConfigSet("url_proxy", "//"+r.Host)
	}

	b := Block{}
	b.mx.Lock()
	defer b.mx.Unlock()

	t1 := time.Now()
	stat := map[string]interface{}{}
	stat["start"] = t1
	stat["status"] = "OK"
	stat["title"] = block.Title
	stat["id"] = block.Id

	// Включаем режим кеширования
	key := ""
	keyParam := ""
	cacheOn, _ := block.Attr("cache", "value")

	//fmt.Println("BaseCache")

	ll := l.ConfigGet("BaseCache")
	if ll != "" && cacheOn != "" && enableCache {

		key, keyParam = l.SetCahceKey(r, block)

		// ПРОВЕРКА КЕША (если есть, отдаем из кеша)
		if res, found := l.СacheGet(key, block, r, page, values, keyParam); found {
			stat["cache"] = "true"
			stat["time"] = time.Since(t1)

			result.result = template.HTML(res)
			result.stat = stat

			return result
		}
	}

	b.Value = map[string]interface{}{}

	// обработка всех странных ошибок
	// ВКЛЮЧИТЬ ПОЗЖЕ!!!!
	//defer func() {
	//	if er := recover(); er != nil {
	//		//ft, err := template.ParseFiles("./upload/control/templates/errors/503.html")
	//		//if err != nil {
	//		//	// l.Logger.Error(err)
	//		//}
	//		//t = template.Must(ft, nil)
	//
	//		result.result = l.ModuleError(er, r)
	//		result.err = err
	//	}
	//}()

	dataSet := make(map[string]interface{})
	dataname := "default" // значение по-умолчанию (будет заменено в потоках)

	tplName, _ := block.Attr("module", "src")
	tquery, _ := block.Attr("tquery", "src")

	//fmt.Println("tplName-long: ", tplName)

	// //////////////////////////////////////////////////////////////////////////////
	// в блоке есть настройки поля расширенного фильтра, который можно добавить в самом блоке
	// дополняем параметры request-a, доп. параметрами, которые переданы через блок
	extfilter, _ := block.Attr("extfilter", "value") // дополнительный фильтр для блока
	dv := []Data{block}

	extfilter = l.DogParse(extfilter, r, &dv, b.Value)
	extfilter = strings.Replace(extfilter, "?", "", -1)

	// парсим переденную строку фильтра
	m, err := url.ParseQuery(extfilter)
	if err != nil {
		// l.Logger.Error(err, "Error parsing extfilter from block.")
	}

	// добавляем в URL переданное значение из настроек модуля
	// если этих значений еще нет (НЕ ЗАМЕНЯЕМ)
	var q url.Values
	for k, v := range m {
		q = r.URL.Query() // Get a copy of the query values.
		if _, found := q[k]; !found {
			q.Add(k, join(v, ",")) // Add a new value to the set. Переводим обратно в строку из массива
		}
	}
	if len(m) != 0 {
		r.URL.RawQuery = q.Encode() // Encode and assign back to the original query.
	}
	// //////////////////////////////////////////////////////////////////////////////
	// //////////////////////////////////////////////////////////////////////////////

	tconfiguration, _ := block.Attr("configuration", "value")
	tconfiguration = strings.Replace(tconfiguration, "  ", "", -1)

	uuid := UUID()
	if values != nil && len(values) != 0 {
		for k, v := range values {
			if _, found := b.Value[k]; !found {
				b.Value[k] = v
			}
		}
	}

	b.Value["Rand"] = uuid[1:6] // переопределяем отдельно для каждого модуля
	b.Value["URL"] = r.URL.Query().Encode()
	b.Value["Prefix"] = "/" + Domain + "/" + l.ConfigGet("path_templates")
	b.Value["Domain"] = Domain
	b.Value["CDN"] = l.ConfigGet("url_fs")
	b.Value["Path"] = ClientPath
	b.Value["Title"] = Title
	b.Value["Form"] = r.Form
	b.Value["RequestURI"] = r.RequestURI
	b.Value["Referer"] = r.Referer()
	b.Value["Profile"] = profile
	b.Value["Cookie"] = r.Cookies()
	b.Value["Request"] = r

	// обработк @-функции в конфигурации
	dv = []Data{block}
	dogParseConfiguration := l.DogParse(tconfiguration, r, &dv, b.Value)

	//fmt.Println(r, "\n")

	// конфигурация без обработки @-функции
	var confRaw map[string]Element
	if tconfiguration != "" {
		err = json.Unmarshal([]byte(tconfiguration), &confRaw)
	}

	// конфигурация с обработкой @-функции
	var conf map[string]Element
	if dogParseConfiguration != "" {
		err = json.Unmarshal([]byte(dogParseConfiguration), &conf)
	}

	if err != nil {
		result.result = l.ModuleError("Error json-format configurations: "+marshal(tconfiguration), r)
		result.err = err
		return result
	}

	// сформировал структуру полученных описаний датасетов
	var source []map[string]string
	if d, found := conf["datasets"]; found {
		err := json.Unmarshal([]byte(marshal(d.Source)), &source)
		if err != nil {
			stat["status"] = "error"
			stat["description"] = fmt.Sprint(err)

			result.result = l.ModuleError(err, r)
			result.err = err
			result.stat = stat
			return result
		}
	}

	//fmt.Println(r, "\n")

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

			//fmt.Println(r, "\n\n")

			ress := l.QueryWorker(queryUID, dataname, source, r)

			//fmt.Println("ress: ", ress)

			dataSet[dataname] = ress
		}

	}

	b.Data = dataSet
	b.Page = page
	b.Configuration = conf
	// b.ConfigurationRaw = confRaw
	b.ConfigurationRaw = tconfiguration
	b.Request = r

	// удаляем лишний путь к файлу, добавленную через консоль
	// СЕКЬЮРНО! Если мы вычитаем текущий путь пользователя, то сможем получить доступ к файлам только текущего проекта
	// иначе необходимо будет авторизоваться и правильный путь (например  /console/gui мы не вычтем)
	// НО ПРОБЛЕМА реиспользования ранее загруженных и настроенных путей к шаблонам.
	//tplName = strings.Replace(tplName, Application["client_path"], ".", -1)

	// НЕ СЕКЬЮРНО!
	// вычитаем не текущий client_path а просто две первых секции из адреса к файлу
	// позволяем получить доступ к ранее загруженным путям шаблонов другим пользоватем с другим префиксом
	// ПО-УМОЛЧАНИЮ (для реиспользования модулей и схем)

	// УБРАНО ПРИ ОТЛАДКЕ VFS
	//sliceMake := strings.Split(tplName, "/")
	//// подоплянем только если длинный путь, если uid то нет
	//if len(sliceMake) >= 3 {
	//	tplName = strings.Join(sliceMake[3:], "/")
	//	tplName = l.ConfigGet("workdir") + "/" + tplName
	//}

	// в режиме отладки пересборка шаблонов происходит при каждом запросе
	//var tmpl *template.Template
	//if debugMode {
	//	if len(tplName) > 0 {
	//		name := path.Base(tplName)
	//		tmpl, _ = template.New(name).Funcs(FuncMap).ParseFiles(tplName)
	//
	//	}
	//	if tmpl != nil {
	//		tmpl.Execute(&c, b)
	//	}
	//} else {
	//	t.ExecuteTemplate(&c, tplName, b)
	//}
	//
	//if tmpl != nil {
	//	result.result = template.HTML(c.String())
	//} else {
	//	result.result = "<center><h3>Ошибка обработки файла шаблона (файл не найден) при генерации блока.</h3></center>"
	//}

	uidModule, found := block.Attr("module", "src")
	if !found {
		result.err = fmt.Errorf("Attribute module not found (value: %s)", uidModule)
	}

	blockBody, err := l.generateBlock(tplName, b, uidModule)
	if err != nil {
		result.err = err
	}

	// чистим от лишних пробелов
	re := regexp.MustCompile("(?m)^\\s+")
	blockBody = re.ReplaceAllString(blockBody, "")

	result.result = template.HTML(blockBody)

	// Включаем режим кеширования
	jj := l.ConfigGet("BaseCache")
	if jj != "" && cacheOn != "" && enableCache {
		//fmt.Println("кэш включен")
		key, keyParam = l.SetCahceKey(r, block)

		//fmt.Println(" Начинаем кешировать")
		// КЕШИРОВАНИЕ данных
		l.CacheSet(key, block, page, blockBody, keyParam)
		// log.Warning("CacheSet: ",fl)
	}

	stat["cache"] = "false"
	stat["time"] = time.Since(t1)
	result.stat = stat

	// докидываем ошибку в ответ, если она есть (для отладки и удобства разработки)
	// TODO прямой вывод ошибку в тело страницы - передалать, очень грубо
	if err != nil {
		result.result = template.HTML(fmt.Sprint(err))
	}

	return result
}

// ДЛЯ ПАРАЛЛЕЛЬНОЙ сборки модуля
// получаем объект модуля (отображения)
func (l *app) ModuleBuildParallel(ctxM context.Context, p Data, r *http.Request, page Data, values map[string]interface{}, enableCache bool, buildChan chan ModuleResult, wg *sync.WaitGroup) {
	defer wg.Done()
	t1 := time.Now()

	result := ModuleResult{}

	// проверка на выход по сигналу
	select {
	case <-ctxM.Done():
		return
	default:
	}

	//fmt.Println("Начали ModuleBuildParallel:", p.Id, p.Title)

	// заменяем в State localhost на адрес домена (если это подпроцесс то все норм, но если это корневой сервис,
	// то у него url_proxy - localhost и узнать реньше адрес мы не можем, ибо еще домен не инициировался
	// а значит подменяем localhost при каждом обращении к модулю
	//if strings.Contains(l.State["url_proxy"], "localhost") {
	//	url_shema := "http"
	//	if r.TLS != nil {
	//		url_shema = "https"
	//	}
	//	l.State["url_proxy"] = url_shema + "://" + r.Host
	//}

	if strings.Contains(l.ConfigGet("url_proxy"), "localhost") {
		//url_shema := "http"
		//if r.TLS != nil {
		//	url_shema = "https"
		//}
		//l.State["url_proxy"] = url_shema + "://" + r.Host
		l.ConfigSet("url_proxy", "//"+r.Host)
	}

	// указатель на профиль текущего пользователя
	ctx := r.Context()
	var profile ProfileData
	profileRaw := ctx.Value("UserRaw")
	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)

	var c bytes.Buffer
	var b Block
	var err error
	var key, keyParam string
	b.Value = map[string]interface{}{}
	result.id = p.Id

	stat := map[string]interface{}{}
	stat["start"] = t1
	stat["status"] = "OK"
	stat["title"] = p.Title
	stat["id"] = p.Id

	//////////////////////////////
	// Включаем режим кеширования
	//////////////////////////////
	cacheOn, _ := p.Attr("cache", "value")

	if l.ConfigGet("BaseCache") != "" && cacheOn != "" && enableCache {

		key, keyParam := l.SetCahceKey(r, p)

		// ПРОВЕРКА КЕША (если есть, отдаем из кеша)
		if res, found := l.СacheGet(key, p, r, page, values, keyParam); found {
			stat["cache"] = "true"
			stat["time"] = time.Since(t1)

			result.result = template.HTML(res)
			result.stat = stat

			buildChan <- result
			return
		}
	}
	//////////////////////////////
	//////////////////////////////

	// проверка на выход по сигналу
	select {
	case <-ctxM.Done():
		return
	default:
	}

	// обработка всех странных ошибок
	//defer func() {
	//	if er := recover(); er != nil {
	//		t = template.Must(template.ParseFiles("./upload/control/templates/errors/503.html"))
	//		result.result = ModuleError(er, r)
	//	}
	//}()

	dataSet := make(map[string]interface{})
	dataname := "default" // значение по-умолчанию (будет заменено в потоках)

	tplName, _ := p.Attr("module", "src")
	tquery, _ := p.Attr("tquery", "src")

	// //////////////////////////////////////////////////////////////////////////////
	// в блоке есть настройки поля расширенного фильтра, который можно добавить в самом блоке
	// дополняем параметры request-a, доп. параметрами, которые переданы через блок
	extfilter, _ := p.Attr("extfilter", "value") // дополнительный фильтр для блока
	dp := []Data{p}
	extfilter = l.DogParse(extfilter, r, &dp, b.Value)
	extfilter = strings.Replace(extfilter, "?", "", -1)

	// парсим переденную строку фильтра
	m, err := url.ParseQuery(extfilter)
	if err != nil {
		// l.Logger.Error(err, "Error parsing extfilter from block.")
	}

	// добавляем в URL переданное значение из настроек модуля
	var q url.Values
	for k, v := range m {
		q = r.URL.Query()      // Get a copy of the query values.
		q.Add(k, join(v, ",")) // Add a new value to the set. Переводим обратно в строку из массива
	}
	if len(m) != 0 {
		r.URL.RawQuery = q.Encode() // Encode and assign back to the original query.
	}
	// //////////////////////////////////////////////////////////////////////////////
	// //////////////////////////////////////////////////////////////////////////////

	tconfiguration, _ := p.Attr("configuration", "value")
	tconfiguration = strings.Replace(tconfiguration, "  ", "", -1)

	uuid := UUID()

	if values != nil && len(values) != 0 {
		for k, v := range values {
			if _, found := b.Value[k]; !found {
				b.Value[k] = v
			}
		}
	}

	b.Value["Rand"] = uuid[1:6] // переопределяем отдельно для каждого модуля
	b.Value["URL"] = r.URL.Query().Encode()
	b.Value["Prefix"] = "/" + Domain + "/" + l.ConfigGet("path_templates")
	b.Value["Domain"] = Domain
	b.Value["CDN"] = l.ConfigGet("url_fs")
	b.Value["Path"] = ClientPath
	b.Value["Title"] = Title
	b.Value["Form"] = r.Form
	b.Value["RequestURI"] = r.RequestURI
	b.Value["Referer"] = r.Referer()
	b.Value["Profile"] = profile
	b.Value["Cookie"] = r.Cookies()
	b.Value["Request"] = r

	// обработк @-функции в конфигурации
	dp = []Data{p}
	dogParseConfiguration := l.DogParse(tconfiguration, r, &dp, b.Value)

	// конфигурация без обработки @-функции
	var confRaw map[string]Element
	if tconfiguration != "" {
		err = json.Unmarshal([]byte(tconfiguration), &confRaw)
	}

	// конфигурация с обработкой @-функции
	var conf map[string]Element
	if dogParseConfiguration != "" {
		err = json.Unmarshal([]byte(dogParseConfiguration), &conf)
	}

	if err != nil {
		result.result = l.ModuleError("Error json-format configurations: "+marshal(tconfiguration), r)
		result.err = err
		buildChan <- result

		//dd := map[string]template.HTML{key:ModuleError("Error json-format configurations: "+marshal(tconfiguration), r)}
		return
	}

	// сформировал структуру полученных описаний датасетов
	var source []map[string]string
	if d, found := conf["datasets"]; found {
		err := json.Unmarshal([]byte(marshal(d.Source)), &source)
		if err != nil {
			result.result = l.ModuleError(err, r)
			buildChan <- result
			return
		}
	}

	// ПЕРЕДЕЛАТЬ НА ПАРАЛЛЕЛЬНЫЕ ПОТОКИ
	if tquery != "" {
		var name, uid string
		slquery := strings.Split(tquery, ",")

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
			ress := l.QueryWorker(queryUID, dataname, source, r)
			dataSet[dataname] = ress
		}
	}

	b.Data = dataSet
	b.Page = page
	b.Metric = Metric
	b.Configuration = conf
	//b.ConfigurationRaw = confRaw
	b.ConfigurationRaw = tconfiguration

	b.Request = r

	// удаляем лишний путь к файлу, добавленную через консоль
	// СЕКЬЮРНО! Если мы вычитаем текущий путь пользователя, то сможем получить доступ к файлам только текущего проекта
	// иначе необходимо будет авторизоваться и правильный путь (например  /console/gui мы не вычтем)
	// НО ПРОБЛЕМА реиспользования ранее загруженных и настроенных путей к шаблонам.
	//tplName = strings.Replace(tplName, Application["client_path"], ".", -1)

	// НЕ СЕКЬЮРНО!
	// вычитаем не текущий client_path а просто две первых секции из адреса к файлу
	// позволяем получить доступ к ранее загруженным путям шаблонов другим пользоватем с другим префиксом
	// ПО-УМОЛЧАНИЮ (для реиспользования модулей и схем)
	//sliceMake := strings.Split(tplName, "/")
	//if strings.Contains(tplName, sep) {
	//	if len(sliceMake) < 3 {
	//		err = errors.New("Error: The path to the module file is incorrect or an error occurred while selecting the module in the block object!")
	//		fmt.Println("начали вышли по ошибке в len(sliceMake)")
	//
	//		l.ErrorModuleBuild(stat, buildChan, time.Since(t1), err)
	//
	//		fmt.Println("вышли по ошибке в len(sliceMake)")
	//		return
	//	}
	//	tplName = strings.Join(sliceMake[3:], "/")
	//	tplName = l.ConfigGet("workdir") + "/" + tplName
	//}

	uidModule, _ := p.Attr("module", "src")
	blockBody, err := l.generateBlock(tplName, b, uidModule)
	if err != nil {
		l.ErrorModuleBuild(stat, buildChan, time.Since(t1), err)
		// l.Logger.Error(errT, "Error generated module.")
		return
	}

	stat["cache"] = "true"
	stat["time"] = time.Since(t1)

	// чистим от лишних пробелов
	re := regexp.MustCompile("(?m)^\\s+")
	blockBody = re.ReplaceAllString(blockBody, "")

	result.result = template.HTML(blockBody)
	result.stat = stat

	// Включаем режим кеширования
	if l.ConfigGet("BaseCache") != "" && cacheOn != "" && enableCache {
		key, keyParam = l.SetCahceKey(r, p)

		// КЕШИРОВАНИЕ данных
		l.CacheSet(key, p, page, c.String(), keyParam)
	}

	//fmt.Println("- ModuleBuildParallel: 5")

	stat["cache"] = "false"
	stat["time"] = time.Since(t1)
	result.stat = stat

	buildChan <- result

	//log.Warning("Stop ", p.Title, "-", time.Since(t1))

	return
}

// генерируем блок в зависимости от переданного пути (или из файла или из данных в объекте)
func (l *app) generateBlock(tplName string, bl Block, uidModule string) (res string, err error) {
	var c bytes.Buffer

	// TODO удалить позже
	// очищаем доп.путей, которые были при использовании файловой структуры
	tplName = strings.Replace(tplName, sep+Domain+sep+"gui", "", 1)

	//fmt.Println("generateBlock:", tplName)

	// если содержится разделитель - значит передан путь к файлу (старая версия) и генерируем из файла
	// иначе берем значение из поля codetpl (новая версия), если пусто, то из поля _filecontent_url
	// (для случаем, когда блок выбрали, но содержимое файла не перенесли в новое поле и оно хрантся в поле автосохранения файла)
	if strings.Contains(tplName, sep) {
		c, err = l.generateBlockFromFile(tplName, bl)
		if err != nil {
			err = fmt.Errorf("%s file:'%s' (%s)", "Error: Generate Module from file is failed!", tplName, err)
			res = fmt.Sprint(err)
		}
	} else {
		var objModule ResponseData

		// запроса модуля (в нем HTML-тело шаблона)
		//objModule, err = b.api.ObjGet(uidModule)
		_, err = l.Curl("GET", "_objs/"+uidModule, "", &objModule, nil)
		if err != nil {
			err = fmt.Errorf("%s (%s)", "Error: Get object Module is failed!", err)
			res = fmt.Sprint(err)
		}
		if len(objModule.Data) == 0 {
			err = fmt.Errorf("%s", "Error: Object Module is null!")
			res = fmt.Sprint(err)
		}

		// если выбрано несколько блоков, их все объединяем в один (очередность случайная)
		htmlCode := ""
		for _, v := range objModule.Data {
			codetpl, _ := v.Attr("_filecontent_module", "value")
			if codetpl == "" {
				codetpl, _ = v.Attr("codetpl", "value")
				if codetpl == "" {
					codetpl, _ = v.Attr("_filecontent_url", "value")
				}
			}
			htmlCode = htmlCode + codetpl
		}

		c, err = l.generateBlockFromField(htmlCode, bl)
	}

	res = c.String()

	return res, err
}

// generateBlockFromFile генерируем блок из файла (для совместимости со старыми модулями)
func (b *app) generateBlockFromFile(tplName string, bl Block) (c bytes.Buffer, err error) {
	var tmpl *template.Template

	sliceMake := strings.Split(tplName, "/")
	if len(sliceMake) < 3 {
		err = fmt.Errorf("%s", "Error: The path to the module file is incorrect or an error occurred while selecting the module in the block object!")
		//b.ErrorModuleBuild(stat, buildChan, time.Since(t1), errT)
		return
	}
	//tplName = strings.Join(sliceMake[3:], "/")
	//tplName = b.config.payload["Workingdir"] + "/" + tplName

	dataFile, _, err := b.vfs.Read(tplName)

	tmpl = template.New(tplName).Funcs(FuncMap)
	t, err = tmpl.Parse(string(dataFile))
	if err != nil {
		err = fmt.Errorf("%s", "Error: Getting path.Base failed! tplName: %s", tplName)
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

	//if len(tplName) > 0 {
	//	name := path.Base(tplName)
	//	if name == "" {
	//		err = fmt.Errorf("%s", "Error: Getting path.Base failed!")
	//		tmpl = nil
	//	} else {
	//		tmpl, err = template.New(name).Funcs(FuncMap).ParseFiles(tplName)
	//		if err != nil {
	//			return
	//		}
	//	}
	//}

	return c, err
}

// генерируем блок из переданного текста
func (b *app) generateBlockFromField(value string, bl Block) (c bytes.Buffer, err error) {
	tmpl, err := template.New("name").Funcs(FuncMap).Parse(value)
	if err != nil {
		return
	}
	err = tmpl.Execute(&c, bl)

	return c, err
}

// вываливаем ошибку при генерации модуля
func (c *app) ErrorModuleBuild(stat map[string]interface{}, buildChan chan ModuleResult, timerRun interface{}, errT error) {
	var result ModuleResult

	stat["cache"] = "false"
	stat["time"] = timerRun
	result.stat = stat
	result.result = template.HTML(fmt.Sprint(errT))
	result.err = errT

	buildChan <- result

	return
}

// queryUID - ид-запроса
func (c *app) QueryWorker(queryUID, dataname string, source []map[string]string, r *http.Request) interface{} {
	//var resp Response

	resp := c.GUIQuery(queryUID, r)

	//switch x := resp1.(type) {
	//case Response:
	//	resp = resp1.(Response)
	//
	//default:
	//	resp.Data = resp1
	//}

	///////////////////////////////////////////
	// Расчет пагенации
	///////////////////////////////////////////

	var m3 Response
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

// вывод ошибки выполнения блока
func (c *app) ErrorPage(err interface{}, w http.ResponseWriter, r *http.Request) {
	p := ErrorForm{
		Err: err,
		R:   *r,
	}
	log.Error(err)

	t = template.Must(template.ParseFiles("./upload/control/templates/errors/500.html"))
	t.Execute(w, p)
}

// вывод ошибки выполнения блока
func (l *app) ModuleError(err interface{}, r *http.Request) template.HTML {
	var c bytes.Buffer

	p := ErrorForm{
		Err: err,
		R:   *r,
	}

	// l.Logger.Error(nil,err)
	//fmt.Println("ModuleError: ", err)

	//wd := l.ConfigGet("workdir")
	//t = template.Must(template.ParseFiles(wd + "/upload/control/templates/errors/503.html"))

	t, err = template.ParseGlob(errorTemplate503)

	t.Execute(&c, p)
	result = template.HTML(c.String())

	return result
}

// отправка запроса на получения данных из интерфейса GUI
// параметры переданные в строке (r.URL) отправляем в теле запроса
func (c *app) GUIQuery(tquery string, r *http.Request) Response {

	var resultInterface interface{}
	var err error
	var dataResp, returnResp Response

	formValues := r.PostForm
	bodyJSON, _ := json.Marshal(formValues)

	// добавляем к пути в запросе переданные в блок параметры ULR-а (возможно там есть параметры для фильтров)
	filters := r.URL.RawQuery
	if filters != "" {
		filters = "?" + filters
	}

	// ФИКС!
	// добавляем еще токен (cookie) текущего пользователя
	// это нужно для случая, если мы вызываем запрос из запроса и кука не передается
	// а если куки нет, то сбрасывается авторизация
	//cookieCurrent, err := r.Cookie("sessionID")
	//token := ""
	//if err == nil {
	//	tokenI := strings.Split(fmt.Sprint(cookieCurrent), "=")
	//	if len(tokenI) > 1 {
	//		token = tokenI[1]
	//	}
	//	if token != "" {
	//		if strings.Contains(filters, "?") {
	//			filters = filters + "&token=" + token
	//		} else {
	//			filters = filters + "?token=" + token
	//		}
	//	}
	//}

	// ЛЮТЫЙ ФИКС
	// не получилось передать в app состояние, поэтому добавляю ранее путь к GUI и если он указан, то отправляю по полному пути
	// дополняем путем до API если не передан вызов внешнего запроса через http://
	if tquery[:4] != "http" {
		resultInterface, err = c.Curl(r.Method, "/query/"+tquery+filters, string(bodyJSON), &dataResp, r.Cookies())
		if err != nil {
			//fmt.Println("Error. Request failed: " + "/query/" + tquery + filters)
			return returnResp
		}
	} else {
		resultInterface, err = c.Curl(r.Method, tquery+filters, string(bodyJSON), &dataResp, r.Cookies())
		if err != nil {
			//fmt.Println("Error. Request failed: " + tquery + filters)
			return returnResp
		}
	}

	//fmt.Println(dataResp)
	//if tquery == "qgoups" {
	//fmt.Println("tquery: ", "/query/"+tquery+filters, "; resultInterface: ", resultInterface)
	//}

	// нам тут нужен Response, но бывают внешние запросы,
	// поэтому если не Response то дописываем в Data полученное тело
	if dataResp.Data != nil {
		returnResp = dataResp
	} else {
		if resultInterface == nil {
			returnResp.Data = []string{} // пустое значение (nil) => пустой массив (по-умолчанию)
		} else {
			returnResp.Data = resultInterface
		}
	}

	var dd ResponseData
	ff, _ := json.Marshal(dd)
	json.Unmarshal(ff, &dd)

	return returnResp
}

// удаляем элемент из слайса
func (p *ResponseData) RemoveData(i int) bool {

	if i < len(p.Data) {
		p.Data = append(p.Data[:i], p.Data[i+1:]...)
	} else {
		//log.Warning("Error! Position invalid (", i, ")")
		return false
	}

	return true
}

////////////////////////////////////////////////////////////////////////////////////////
/////////////// ФУНКЦИИ ДЛЯ ВЛОЖЕНИЯ ОБЪЕКТОВ Data В ФОРМАТ ДЕРЕВА /////////////////////
////////////////////////////////////////////////////////////////////////////////////////
// формируем вложенную структуру объектов
func DataToIncl(objData []Data) []*DataTree {

	// переводим slice в map, чтобы можно было удалять объект и обращаться по ключу при формировании подуровней навигатора
	mapLevel := map[string]*DataTree{}
	for _, v := range objData {
		item := DataTree{}

		item.Uid = v.Uid
		item.Source = v.Source
		item.Type = v.Type
		item.Attributes = v.Attributes
		item.Title = v.Title
		item.Type = v.Type
		item.Parent = v.Parent
		item.Rev = v.Rev
		item.Сopies = v.Сopies

		mapLevel[v.Uid] = &item
	}

	// делаю обратное наследование, добавляю в Sub значения всех потомков (для оптимальной функции вложения)
	for _, v := range mapLevel {
		if _, found := v.Attributes["leader"]; found {
			Leader := v.Attributes["leader"].Src
			if Leader != "" && v.Uid != "" {
				d, f := mapLevel[Leader]
				if f {
					d.Sub = append(d.Sub, v.Uid)
				}
			}
		}

	}

	// пробегаем карту полигонов и переносим вложенные внутрь
	for _, item := range mapLevel {
		item.ScanSub(&mapLevel)
	}

	// преобразуем карту в слайс
	sliceNavigator := []*DataTree{}
	for _, m := range mapLevel {
		sliceNavigator = append(sliceNavigator, m)
		//log.Error("============")
		//log.Error(*m)
	}

	// сортируем по order как число
	SortItems(sliceNavigator, "order", "int")

	return sliceNavigator
}

// метод типа Items (перемещаем структуры в карте, исходя из заявленной вложенности элементов)
// (переделать дубль фукнции)
func (p *DataTree) ScanSub(maps *map[string]*DataTree) {
	if p.Sub != nil && len(p.Sub) != 0 {
		for _, c := range p.Sub {
			gg := *maps
			fromP := gg[c]
			if fromP != nil {
				copyPolygon := *fromP
				p.Incl = append(p.Incl, &copyPolygon)
				delete(*maps, c)
				copyPolygon.ScanSub(maps)
			}
		}
	}
}

// сортируем в слейсе полигонов по полю sort
// typesort - тип сортировки (string/int) - если int то преобразуем в число перед сортировкой
// fieldsort - поле для сортировки
func SortItems(p []*DataTree, fieldsort string, typesort string) {

	sort.Slice(p, func(i, j int) bool {

		value1 := "0"
		value2 := "0"
		if typesort == "int" {
			value1 = "0"
			value2 = "0"
		}

		if oi, found := p[i].Attributes[fieldsort]; found {
			if oi.Value != "" {
				value1 = oi.Value
			}
		}
		if oj, found := p[j].Attributes[fieldsort]; found {
			if oj.Value != "" {
				value2 = oj.Value
			}
		}

		vi1, err1 := strconv.Atoi(value1)
		vi2, err2 := strconv.Atoi(value2)

		// если передан int, но произошла ошибка, то не не меняем
		if typesort == "int" {
			if err1 == nil && err2 == nil {
				return vi1 < vi2
			} else {
				return false
			}
		} else {
			// если стринг, то всегда проверяем как-будто это сравнение строк
			return vi1 < vi2
		}

	})

	for i, _ := range p {
		if p[i].Incl != nil && len(p[i].Incl) != 0 {
			f := p[i].Incl
			SortItems(f, fieldsort, typesort)
		}
	}
}

// вспомогательная фукнция выбирает только часть дерево от заданного лидера
func TreeShowIncl(in []*DataTree, obj string) (out []*DataTree) {
	if obj == "" {
		return in
	}

	for _, v := range in {

		if v.Source == obj {
			out = append(out, v)
			return out
		} else {

			out = TreeShowIncl(v.Incl, obj)
			if len(out) != 0 {
				return out
			}
		}

	}
	return out
}
