package app_lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/smtp"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/api-client"
	"git.lowcodeplatform.net/fabric/models"
	"github.com/Masterminds/sprig"
	"github.com/satori/go.uuid"

	"git.lowcodeplatform.net/fabric/lib"
)

var FuncMapS = sprig.FuncMap()

var FuncMap = template.FuncMap{
	"separator":           separator,
	"cookie":              cookieGet,
	"cookieget":           cookieGet,
	"cookieset":           cookieSet,
	"attr":                attr,
	"addfloat":            addfloat,
	"mulfloat":            mulfloat,
	"datetotext":          datetotext,
	"output":              output,
	"cut":                 cut,
	"concatination":       concatination,
	"join":                join,
	"rand":                randT,
	"uuid":                UUID,
	"refind":              refind,
	"rereplace":           rereplace,
	"replace":             Replace,
	"contains":            contains,
	"contains1":           contains1,
	"dict":                dict,
	"dictstring":          dictstring,
	"sum":                 sum,
	"split":               split,
	"splittostring":       splittostring,
	"set":                 set,
	"get":                 get,
	"delete":              deletekey,
	"slicenew":            slicenew,
	"sliceset":            sliceset,
	"sliceappend":         sliceappend,
	"slicedelete":         slicedelete,
	"slicestringnew":      slicestringnew,
	"slicestringset":      slicestringset,
	"slicestringappend":   slicestringappend,
	"slicestringdelete":   slicestringdelete,
	"marshal":             marshal,
	"value":               value,
	"hash":                hash,
	"unmarshal":           unmarshal,
	"compare":             compare,
	"totree":              totree,
	"tostring":            tostring,
	"toint":               toint,
	"tofloat":             tofloat,
	"tointerface":         tointerface,
	"tohtml":              tohtml,
	"timefresh":           Timefresh,
	"timenow":             timenow,
	"timeformat":          timeformat,
	"timetostring":        timetostring,
	"timeyear":            timeyear,
	"timemount":           timemount,
	"timeday":             timeday,
	"timeparse":           timeparse,
	"tomoney":             tomoney,
	"invert":              invert,
	"substring":           substring,
	"dogparse":            dogparse,
	"confparse":           confparse,
	"varparse":            parseparam,
	"parseparam":          parseparam,
	"parseformsep":        parseformsep,
	"divfloat":            divfloat,
	"sendmail":            Sendmail,
	"jsonescape":          jsonEscape,
	"jsonescapeunlessamp": jsonEscapeUnlessAmp,
	"apiobjget":           ObjGet,
	"apiobjcreate":        ObjCreate,
	"apiobjattrupdate":    ObjAttrUpdate,
	"apiobjdelete":        ObjDelete,
	"apilinkget":          LinkGet,
	"apiquery":            Query,
	"curl":                curl,
	"parseform":           parseform,
	"parsebody":           parsebody,
	"redirect":            redirect,
	"groupbyfield":        groupbyfield,
	"sortbyfield":         sortbyfield,

	"randstringslice": randstringslice,

	"objFromID": objFromID,
}

// randomizer перемешивает полученный слайс
func randstringslice(a []string) (res []string) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(a),
		func(i, j int) { a[i], a[j] = a[j], a[i] })

	return a
}

// objFromID получить объект из массива объектов по id
func objFromID(dt []models.Data, id string) (result interface{}) {
	//var dt []models.Data

	//err := json.Unmarshal([]byte(fmt.Sprint(data)), &dt)
	//if err != nil {
	//	return fmt.Sprint(err, data)
	//}
	for _, v := range dt {
		if v.Id == id {
			return v
		}
	}

	return "nil"
}

// groupbyfield группируем полученные данные (объекты) формата models.ResponseData согласно шаблону
func groupbyfield(queryData interface{}, groupField, groupPoint string, sortField, sortPoint string, desc bool) (result interface{}, err error) {
	var d models.ResponseData

	b, _ := json.Marshal(queryData)
	err = json.Unmarshal(b, &d)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal queryData. err: %s, payload: %s", err, string(b))
	}

	// разбираем структуру ответа по заданному полю/поинту
	resultMap := map[string][]models.Data{}
	for _, v := range d.Data {
		key, found := v.Attr(groupField, groupPoint)
		if !found {
			resultMap[key] = []models.Data{}
		}
		resultMap[key] = append(resultMap[key], v)
	}

	// сортируем внутри каждого группы по заданному полю
	resultSortedMap := map[string][]models.Data{}
	for n, m := range resultMap {
		sort.Slice(m, func(i, j int) bool {
			switch sortPoint {
			case "src":
				return m[i].Attributes[sortField].Src < m[j].Attributes[sortField].Src
			case "rev":
				return m[i].Attributes[sortField].Rev < m[j].Attributes[sortField].Rev
			default:
				return m[i].Attributes[sortField].Value < m[j].Attributes[sortField].Value
			}
		})
		resultSortedMap[n] = m
	}

	return resultSortedMap, nil
}

// sortbyfield сортирует (объекты) формата models.ResponseData согласно шаблону
func sortbyfield(queryData interface{}, sortField, sortPoint string, desc bool) (result interface{}, err error) {
	var d models.ResponseData

	b, _ := json.Marshal(queryData)
	err = json.Unmarshal(b, &d)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal queryData. err: %s, payload: %s", err, string(b))
	}

	// сортируем внутри каждого группы по заданному полю
	m := d.Data
	sort.Slice(m, func(i, j int) bool {
		switch sortPoint {
		case "src":
			return m[i].Attributes[sortField].Src < m[j].Attributes[sortField].Src
		case "rev":
			return m[i].Attributes[sortField].Rev < m[j].Attributes[sortField].Rev
		default:
			return m[i].Attributes[sortField].Value < m[j].Attributes[sortField].Value
		}
	})

	d.Data = m

	return d, nil
}

// redirect
func redirect(w http.ResponseWriter, r *http.Request, url string, statusCode int) (result interface{}, err error) {
	http.Redirect(w, r, url, statusCode)

	return
}

// parsebody
// json - преобразованный, по-умолчанию raw (строка)
func parsebody(r *http.Request, format string) (result interface{}, err error) {
	responseData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	switch format {
	case "json":
		err = json.Unmarshal(responseData, &result)
		if err != nil {
			return string(responseData), fmt.Errorf("error Unmarshal (output data in raw format. err: %s, raw: %s", err, string(responseData))
		}
	default:
		return string(responseData), nil
	}

	return
}

// parseform парсит полученные в запросе значения в мапку
func parseformsep(r *http.Request, separator string) map[string]string {
	if separator == "" {
		separator = ","
	}
	err := r.ParseForm()
	if err != nil {
		return nil
	}
	data := make(map[string]string)
	for k, v := range r.PostForm {
		data[k] = strings.Join(v, separator)
	}
	return data
}

func parseform(r *http.Request) map[string][]string {
	err := r.ParseForm()
	if err != nil {
		return nil
	}
	data := make(map[string][]string)
	for k, v := range r.PostForm {
		data[k] = v
	}
	return data
}

// отправка запроса (повторяет интерфейс lib.Curl)
// при передаче в куку времени протухани используется формат "2006-01-02T15:04:05Z07:00"
func curl(method, urlc, bodyJSON string, headers map[string]interface{}, incookies []map[string]interface{}) (rawPayload interface{}) {
	var cookies []*http.Cookie
	var responseData models.ResponseData

	if len(incookies) != 0 {
		for _, v := range incookies {
			c := http.Cookie{}
			for i, j := range v {
				switch strings.ToUpper(i) {
				case "NAME":
					c.Name = fmt.Sprint(j)
				case "VALUE":
					c.Value = fmt.Sprint(j)
				case "PATH":
					c.Path = fmt.Sprint(j)
				case "DAMAIN":
					c.Domain = fmt.Sprint(j)
				case "EXPIRES":
					c.Expires, _ = time.Parse(time.RFC3339, fmt.Sprint(j))
				}
			}
			cookies = append(cookies, &c)
		}
	}

	// КОСТЫЛЬ
	// приводим к формату [string]string - потому, что unmarshal возвращаем [string]interface
	h := map[string]string{}
	for k, v := range headers {
		h[k] = fmt.Sprint(v)
	}

	raw, err := lib.Curl(method, urlc, bodyJSON, responseData, h, cookies)
	if len(responseData.Data) == 0 {
		if err != nil {
			return fmt.Sprint(err)
		}
		return raw
	}

	return responseData
}

// ObjGet операции с объектами через клиента API
func ObjGet(apiURL string, uids string) (result models.ResponseData, err error) {
	ctx := context.Background()
	return api.New(ctx, apiURL, true, 3, 5*time.Second, 5*time.Second).ObjGet(ctx, uids)
}

func ObjCreate(apiURL string, bodymap map[string]string) (result models.ResponseData, err error) {
	ctx := context.Background()
	return api.New(ctx, apiURL, true, 3, 5*time.Second, 5*time.Second).ObjCreate(ctx, bodymap)
}

func ObjDelete(apiURL string, uids string) (result models.ResponseData, err error) {
	ctx := context.Background()
	return api.New(ctx, apiURL, true, 3, 5*time.Second, 5*time.Second).ObjDelete(ctx, uids)
}

func ObjAttrUpdate(apiURL string, uid, name, value, src, editor string) (result models.ResponseData, err error) {
	ctx := context.Background()
	return api.New(ctx, apiURL, true, 3, 5*time.Second, 5*time.Second).ObjAttrUpdate(ctx, uid, name, value, src, editor)
}

func LinkGet(apiURL string, tpl, obj, mode, short string) (result models.ResponseData, err error) {
	ctx := context.Background()
	return api.New(ctx, apiURL, true, 3, 5*time.Second, 5*time.Second).LinkGet(ctx, tpl, obj, mode, short)
}

func Query(apiURL string, query, method, bodyJSON string) (result string, err error) {
	ctx := context.Background()
	return api.New(ctx, apiURL, true, 3, 5*time.Second, 5*time.Second).Query(ctx, query, method, bodyJSON)
}

// формируем сепаратор для текущей ОС
func separator() string {
	return string(filepath.Separator)
}

// Sendmail отправка email-сообщения
// from - от кого отправляется <petrov@mail.ru> или [petrov@mail.ru] или Петров [petrov@mail.ru] или Петров <petrov@mail.ru>
// to - кому (можно несколько через запятую)
// server - почтовый сервер
// port - порт сервера (число в текстовом виде)
// user - пользователь почтового сервера
// pass - пароль пользователя
// message - сообщение
// turbo - режим отправки в отдельной горутине
func Sendmail(server, port, user, pass, from, to, subject, message, turbo string) (result string) {
	var resMessage interface{}
	var fromFull, toFull, subjectFull string
	result = "true"

	f := func() {

		auth := smtp.PlainAuth("", user, pass, server)

		// приводим к одному виду чтобы можно было использвоать и <> и []
		from = Replace(from, ">", "]", -1)
		from = Replace(from, "<", "[", -1)
		to = Replace(to, ">", "]", -1)
		to = Replace(to, "<", "[", -1)

		slFrom := strings.Split(from, ",")
		slTo := strings.Split(to, ",")

		addrFrom := []string{}
		for _, v := range slFrom {
			addr := ""
			a1 := strings.Split(v, "[")
			if len(a1) == 1 { // нет имени, только адрес
				addr = strings.TrimSpace(a1[0])
			} else {
				addr = strings.Trim(a1[1], "]")
			}
			addrFrom = append(addrFrom, addr)
		}

		addrTo := []string{}
		for _, v := range slTo {
			addr := ""
			a1 := strings.Split(v, "[")
			if len(a1) == 1 { // нет имени, только адрес
				addr = strings.TrimSpace(a1[0])
			} else {
				addr = strings.Trim(a1[1], "]")
			}
			addrTo = append(addrTo, addr)
		}

		from = Replace(from, "]", ">", -1)
		from = Replace(from, "[", "<", -1)
		to = Replace(to, "]", ">", -1)
		to = Replace(to, "[", "<", -1)
		mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		fromFull = "From: " + from + "\n"
		toFull = "To: " + to + "\n"
		subjectFull = "Subject: " + subject + "\n"

		resMessage = message

		if len(message) > 5 {
			if message[:4] == "http" {
				fmt.Println("запрос на message: ", message)
				resp, err := http.Get(message)
				responseData, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					resMessage = string(responseData)
				}
			}
		}

		sendMes := subjectFull + fromFull + toFull + mime + fmt.Sprint(resMessage)
		fmt.Println(sendMes)
		if err := smtp.SendMail(server+":"+port, auth, join(addrFrom, ","), addrTo, []byte(sendMes)); err != nil {
			result = fmt.Sprintln(err)
		}
	}

	if turbo == "on" || turbo == "true" {
		go f()
	} else {
		f()
	}

	fmt.Println("Email Sent! - ", result)

	return result
}

func divfloat(a, b interface{}) interface{} {
	aF := fmt.Sprint(a)
	bF := fmt.Sprint(b)
	fa, err := strconv.ParseFloat(aF, 64)
	fb, err := strconv.ParseFloat(bF, 64)

	if err != nil {
		return nil
	}

	return fa / fb
}

// обработка @-функций внутри конфигурации (в шаблонизаторе)
func confparse(configuration string, r *http.Request, queryData interface{}) (result interface{}) {
	var d Data
	var lb app

	b, err := json.Marshal(queryData)
	json.Unmarshal(b, &d)

	if err != nil {
		return "Error! Failed marshal queryData: " + fmt.Sprint(err)
	}
	dv := []Data{d}
	confParse := lb.DogParse(configuration, r, &dv, nil)

	// конфигурация с обработкой @-функции
	var conf map[string]Element
	if confParse != "" {
		err = json.Unmarshal([]byte(confParse), &conf)
	}

	if err != nil {
		return "Error! Failed unmarshal confParse: " + fmt.Sprint(err) + " - " + confParse
	}

	return conf
}

// обработка @-функций внутри шаблонизатора
func dogparse(p string, r *http.Request, queryData interface{}, values map[string]interface{}) (result string) {
	var d Data
	var lb app

	b, _ := json.Marshal(queryData)
	json.Unmarshal(b, &d)

	dv := []Data{d}
	result = lb.DogParse(p, r, &dv, values)

	return result
}

// отдаем значение куки
func cookieGet(name string, field string, r *http.Request) (result string) {
	c, err := r.Cookie(name)
	if err != nil {
		return fmt.Sprint(err)
	}

	field = strings.ToUpper(field)
	switch field {
	case "EXPIRES":
		result = c.Expires.String()
	case "MAXAGE":
		result = fmt.Sprint(c.MaxAge)
	case "PATH":
		result = c.Path
	case "SECURE":
		result = fmt.Sprint(c.Secure)
	default:
		result = c.Value
	}

	return result
}

func cookieSet(name string, field string, r *http.Request) (err error) {
	return nil
}

// получаем значение из переданного объекта
func attr(name, element string, data interface{}) (result interface{}) {
	var dt Data
	var err error
	err = json.Unmarshal([]byte(marshal(data)), &dt)
	if err != nil {
		return fmt.Sprintf("error Unmarshal (func attr). err: %s", err)
	}

	dtl := &dt
	result, _ = dtl.Attr(name, element)

	return result
}

// получаем значение из переданного объекта
func addfloat(i ...interface{}) (result float64) {
	var a float64 = 0.0
	for _, b := range i {
		a += tofloat(b)
	}
	return a
}

func UUID() string {
	stUUID := uuid.NewV4()
	return stUUID.String()
}

func randT() string {
	uuid := UUID()
	return uuid[1:6]
}

func timenow() time.Time {
	return time.Now().UTC()
}

// преобразуем текст в дату (если ошибка - возвращаем false), а потом обратно в строку нужного формата
func timeformat(str interface{}, mask, format string) string {
	ss := fmt.Sprintf("%v", str)

	timeFormat, err := timeparse(ss, mask)
	if err != nil {
		return fmt.Sprint(err)
	}
	res := timeFormat.Format(format)

	return res
}

// преобразуем текст в дату (если ошибка - возвращаем false), а потом обратно в строку нужного формата
func timetostring(time time.Time, format string) string {
	res := time.Format(format)

	return res
}

func timeyear(t time.Time) string {
	ss := fmt.Sprintf("%v", t.Year())
	return ss
}

func timemount(t time.Time) string {
	ss := fmt.Sprintf("%v", t.Month())
	return ss
}

func timeday(t time.Time) string {
	ss := fmt.Sprintf("%v", t.Day())
	return ss
}

func timeparse(str, mask string) (res time.Time, err error) {
	mask = strings.ToUpper(mask)

	time.Now().UTC()
	switch mask {
	case "UTC":
		res, err = time.Parse("2006-02-01 15:04:05 -0700 UTC", str)
	case "NOW", "THIS":
		res, err = time.Parse("2006-02-01 15:04:05", str)
	case "ANSIC":
		res, err = time.Parse(time.ANSIC, str)
	case "UNIXDATE":
		res, err = time.Parse(time.UnixDate, str)
	case "RUBYDATE":
		res, err = time.Parse(time.RubyDate, str)
	case "RFC822":
		res, err = time.Parse(time.RFC822, str)
	case "RFC822Z":
		res, err = time.Parse(time.RFC822Z, str)
	case "RFC850":
		res, err = time.Parse(time.RFC850, str)
	case "RFC1123":
		res, err = time.Parse(time.RFC1123, str)
	case "RFC1123Z":
		res, err = time.Parse(time.RFC1123Z, str)
	case "RFC3339":
		res, err = time.Parse(time.RFC3339, str)
	case "RFC3339NANO":
		res, err = time.Parse(time.RFC3339Nano, str)
	case "STAMP":
		res, err = time.Parse(time.Stamp, str)
	case "STAMPMILLI":
		res, err = time.Parse(time.StampMilli, str)
	case "STAMPMICRO":
		res, err = time.Parse(time.StampMicro, str)
	case "STAMPNANO":
		res, err = time.Parse(time.StampNano, str)
	default:
		res, err = time.Parse(mask, str)
	}

	return res, err
}

func refind(mask, str string, n int) (res [][]string) {
	if n == 0 {
		n = -1
	}
	re := regexp.MustCompile(mask)
	res = re.FindAllStringSubmatch(str, n)

	return
}

func rereplace(str, mask, new string) (res string) {
	re := regexp.MustCompile(mask)
	res = re.ReplaceAllString(str, new)

	return
}

func parseparam(str string, configuration, data interface{}, resulttype string) (result interface{}) {

	// разбиваем строку на слайс для замкены и склейки
	sl := strings.Split(str, "%")

	if len(sl) > 1 {
		for k, v := range sl {
			// нечетные значения - это выделенные переменные
			if k == 1 || k == 3 || k == 5 || k == 7 || k == 9 || k == 11 || k == 13 || k == 15 || k == 17 || k == 19 || k == 21 || k == 23 {
				resfunc := output(v, configuration, data, resulttype)
				sl[k] = fmt.Sprint(resfunc)
			}
		}
		result = strings.Join(sl, "")
	} else {
		result = str
	}

	if resulttype == "html" {
		result = template.HTML(result.(string))
	}

	return result
}

// функция указывает переданная дата до или после текущего времени
func Timefresh(str interface{}) string {

	ss := fmt.Sprintf("%v", str)
	start := time.Now().UTC()

	format := "2006-01-02 15:04:05 +0000 UTC"
	end, _ := time.Parse(format, ss)

	if start.After(end) {
		return "true"
	}

	return "false"
}

// инвертируем строку
func invert(str string) string {
	var result string
	for i := len(str); i > 0; i-- {
		result = result + string(str[i-1])
	}
	return result
}

// переводим массив в строку
func join(slice []string, sep string) (result string) {
	result = strings.Join(slice, sep)

	return result
}

// split разбиваем строку на строковый слайс
func split(str, sep string) (result interface{}) {
	result = strings.Split(str, sep)

	return result
}

// split разбиваем строку на строковый слайс
func splittostring(str, sep string) (result []string) {
	result = strings.Split(str, sep)

	return result
}

// переводим в денежное отображение строки - 12.344.342
func tomoney(str, dec string) (res string) {

	for i, v1 := range invert(str) {
		if (i == 3) || (i == 6) || (i == 9) {
			if (string(v1) != " ") && (string(v1) != "+") && (string(v1) != "-") {
				res = res + dec
			}
		}
		res = res + string(v1)
	}
	return invert(res)
}

func contains1(message, str, substr string) string {
	sl1 := strings.Split(substr, "|")
	for _, v := range sl1 {
		if strings.Contains(str, v) {
			return message
		}
	}
	return ""
}

func contains(str, substr, message, messageelse string) string {
	sl1 := strings.Split(substr, "|")
	for _, v := range sl1 {
		if strings.Contains(str, v) {
			return message
		}
	}
	return messageelse
}

// преобразую дату из 2013-12-24 в 24 января 2013
func datetotext(str string) (result string) {
	mapMount := map[string]string{"01": "января", "02": "февраля", "03": "марта", "04": "апреля", "05": "мая", "06": "июня", "07": "июля", "08": "августа", "09": "сентября", "10": "октября", "11": "ноября", "12": "декабря"}
	spd := strings.Split(str, "-")
	if len(spd) == 3 {
		result = spd[2] + " " + mapMount[spd[1]] + " " + spd[0]
	} else {
		result = str
	}

	return result
}

// заменяем
func Replace(str, old, new string, n int) (message string) {
	message = strings.Replace(str, old, new, n)
	return message
}

// сравнивает два значения и вы	водит текст, если они равны
func compare(var1, var2, message string) string {
	if var1 == var2 {
		return message
	}
	return ""
}

// фукнцкия мультитпликсирования передаваемых параметров при передаче в шаблонах нескольких параметров
// {{template "sub-template" dict "Data" . "Values" $.Values}}
func dict(values ...interface{}) (map[string]interface{}, error) {
	if len(values) == 0 {
		return nil, errors.New("invalid dict call")
	}

	dict := make(map[string]interface{})

	for i := 0; i < len(values); i++ {
		key, isset := values[i].(string)
		if !isset {
			if reflect.TypeOf(values[i]).Kind() == reflect.Map {
				m := values[i].(map[string]interface{})
				for i, v := range m {
					dict[i] = v
				}
			} else {
				return nil, errors.New("dict values must be maps")
			}
		} else {
			i++
			if i == len(values) {
				return nil, errors.New("specify the key for non array values")
			}
			dict[key] = values[i]
		}

	}
	return dict, nil
}

func dictstring(values ...string) (map[string]string, error) {
	if len(values) == 0 {
		return nil, errors.New("invalid dict call")
	}
	if len(values)%2 != 0 {
		return nil, errors.New("params is not even")
	}

	dict := make(map[string]string)
	for i := 0; i < len(values); i = i + 2 {
		dict[values[i]] = values[i+1]
	}
	return dict, nil
}

// slicestringset - заменяем значение в переданном слайсе
func slicestringset(d []string, index int, value string) []string {
	d[index] = value
	return d
}

// slicenew - создаем строковый слайс интерфейсов
func slicenew() []interface{} {
	return []interface{}{}
}

// slicestringnew - создаем строковый слайс
func slicestringnew() []string {
	return []string{}
}

// sliceset - заменяем значение в переданном слайсе
func sliceset(d []interface{}, index int, value interface{}) []interface{} {
	d[index] = value
	return d
}

// sliceappend - добавляет значение в переданный слайс
func sliceappend(d []interface{}, value interface{}) []interface{} {
	return append(d, value)
}

// sliceappend - добавляет значение в переданный слайс
func slicestringappend(d []string, value string) []string {
	return append(d, value)
}

// slicedelete - удаляет из слайса
func slicedelete(d []interface{}, index int) []interface{} {
	copy(d[index:], d[index+1:])
	d[len(d)-1] = ""
	d = d[:len(d)-1]

	return d
}

// slicestringdelete - удаляет из слайса
func slicestringdelete(d []string, index int) []string {
	copy(d[index:], d[index+1:])
	d[len(d)-1] = ""
	d = d[:len(d)-1]

	return d
}

func set(d map[string]interface{}, key string, value interface{}) map[string]interface{} {
	d[key] = value
	return d
}

func get(d map[string]interface{}, key string) (value interface{}) {
	value, found := d[key]
	if !found {
		value = ""
	}
	return value
}

func deletekey(d map[string]interface{}, key string) (value string) {
	delete(d, key)
	return "true"
}

// суммируем
func sum(res, i int) int {
	res = res + i
	return res
}

// образаем по заданному кол-ву символов
func cut(res string, i int, sep string) string {
	res = strings.Trim(res, " ")
	if i <= len([]rune(res)) {
		res = string([]rune(res)[:i]) + sep
	}
	return res
}

// обрезаем строку (строка, откуда, [сколько])
// если откуда отрицательно, то обрезаем с конца
func substring(str string, args ...int) string {
	str = strings.Trim(str, " ")
	lenstr := len([]rune(str))
	from := 0
	count := 0

	// разобрали параметры
	for i, v := range args {
		if i == 0 {
			from = v
		}
		if i == 1 {
			count = v
		}
	}

	to := from + count

	if to > lenstr {
		to = lenstr
	}

	// длина строки меньше чем ДО куда надо образать
	if from < 0 {
		return string([]rune(str)[lenstr+from:]) // с конца
	}

	if count == 0 {
		return string([]rune(str)[from:]) // вырежем все символы до конца строки
	}

	return string([]rune(str)[from:to]) // вырежем диапазон
}

func tostring(i interface{}) (res string) {
	res = fmt.Sprint(i)
	return res
}

// функция преобразует переданные данные в формат типа Items с вложенными подпукнтами
func totree(i interface{}, objstart string) (res interface{}) {
	var objD []Data
	var objRes []interface{}
	var objTree []DataTreeOut

	b3, _ := json.Marshal(i)
	err := json.Unmarshal(b3, &objD)
	if err != nil {
		return "Error convert to ResponseData"
	}

	in := DataToIncl(objD)
	resTree := TreeShowIncl(in, objstart)

	// наполняем дерево
	for _, v := range resTree {
		objRes = append(objRes, v)
	}

	b, _ := json.Marshal(objRes)
	err = json.Unmarshal(b, &objTree)
	if err != nil {
		return "Error convert to DataTree"
	}

	//c, _ := json.Marshal(objRes)
	//res = string(c)

	return objTree
}

func tohtml(i interface{}) template.HTML {

	return template.HTML(i.(string))
}

func toint(i interface{}) (res int) {
	str := fmt.Sprint(i)
	i = strings.Trim(str, " ")
	res, err := strconv.Atoi(str)
	if err != nil {
		return -1
	}

	return res
}

// mulfloat умножение с запятой
func mulfloat(a float64, v ...float64) float64 {
	for _, b := range v {
		a = a * b
	}

	return a
}

func tofloat(i interface{}) (res float64) {
	str := fmt.Sprint(i)
	str = strings.Trim(str, " ")
	str = strings.ReplaceAll(str, ",", ".")
	res, e := strconv.ParseFloat(str, 10)
	if e != nil {
		return -1
	}

	return res
}

func tointerface(input interface{}) (res interface{}) {
	b3, _ := json.Marshal(input)
	err := json.Unmarshal(b3, &res)
	if err != nil {
		return err
	}
	return
}

func concatination(values ...string) (res string) {
	res = strings.Join(values, "")
	return res
}

func marshal(i interface{}) (res string) {
	b3, _ := json.Marshal(&i)
	return string(b3)
}

func unmarshal(i string) (res interface{}) {
	var conf interface{}
	i = strings.Trim(i, "  ")

	err := json.Unmarshal([]byte(i), &conf)
	if err != nil {
		return err
	}
	return conf
}

// СТАРОЕ! ДЛЯ РАБОТЫ В ШАБЛОНАХ ГУЯ СТАРЫХ (ПЕРЕДЕЛАТЬ И УБРАТЬ)
// получаем значение из массива данных по имени элемента
// ПЕРЕДЕЛАТЬ! приходится постоянно сериализовать данные
func value(element string, configuration, data interface{}) (result interface{}) {
	var conf map[string]Element
	json.Unmarshal([]byte(marshal(configuration)), &conf)

	var dt Data
	json.Unmarshal([]byte(marshal(data)), &dt)

	if element == "" {
		return
	}
	if conf == nil {
		return conf
	}

	for k, v := range conf {
		if k == element {

			if v.Type == "text" {
				result = v.Source
			}

			if v.Type == "element" {
				var source map[string]string

				json.Unmarshal([]byte(marshal(v.Source)), &source)
				field := ""
				point := ""
				if _, found := source["field"]; found {
					field = source["field"]
				}
				if _, found := source["point"]; found {
					point = source["point"]
				}

				result, _ = dt.Attr(field, point)
			}

			if v.Type == "object" {
				result, _ = dt.Attr(v.Source.(string), v.Source.(string))
			}

		}
	}

	return result
}

// получаем значение из массива данных по имени элемента
// ПЕРЕДЕЛАТЬ! приходится постоянно сериализовать данные
func output(element string, configuration, data interface{}, resulttype string) (result interface{}) {
	var conf map[string]Element
	json.Unmarshal([]byte(marshal(configuration)), &conf)

	if resulttype == "" {
		resulttype = "text"
	}

	var dt Data
	json.Unmarshal([]byte(marshal(data)), &dt)

	if element == "" {
		return ""
	}
	if conf == nil {
		return ""
	}

	for k, v := range conf {
		if k == element {

			if v.Type == "text" {
				result = v.Source
			}

			if v.Type == "structure" {
				result = v.Source
			}

			if v.Type == "element" {
				var source map[string]string

				json.Unmarshal([]byte(marshal(v.Source)), &source)
				field := ""
				point := ""
				if _, found := source["field"]; found {
					field = source["field"]
				}
				if _, found := source["point"]; found {
					point = source["point"]
				}

				result, _ = dt.Attr(field, point)
			}

			if v.Type == "object" {
				//var args interface{}
				//json.Unmarshal([]byte(marshal(v.Source)), &args)

				result = v.Source
			}

		}
	}

	if resulttype == "html" {
		result = template.HTML(result.(string))
	}

	return result
}

// экранируем "
// fmt.Println(jsonEscape(`dog "fish" cat`))
// output: dog \"fish\" cat
func jsonEscape(i string) (result string) {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)

	return s[1 : len(s)-1]
}

// экранируем кроме аперсанда (&)
func jsonEscapeUnlessAmp(s string) (result string) {
	result = jsonEscape(s)
	result = strings.Replace(result, `\u0026`, "&", -1)
	return result
}

func hash(str string) string {
	var lb *app

	return lb.hash(str)
}
