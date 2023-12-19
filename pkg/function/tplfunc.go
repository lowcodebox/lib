package function

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	api "git.lowcodeplatform.net/fabric/api-client"
	applib "git.lowcodeplatform.net/fabric/app/lib"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/app/pkg/tree"
	"git.lowcodeplatform.net/fabric/models"
	"git.lowcodeplatform.net/packages/logger"
	"github.com/Masterminds/sprig"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"go.uber.org/zap"
)

var FuncMapS = sprig.FuncMap()

type tplfunc struct {
	cfg  model.Config
	tree tree.Tree
	api  api.Api
}

type TplFunc interface {
	Cookie(name string, field string, r *http.Request) (result string)
	GetFuncMap() template.FuncMap
	Separator() string
	Sendmail(server, port, user, pass, from, to, subject, message, turbo string) (result string)
	Divfloat(a, b interface{}) float64
	Mulfloat(a float64, v ...float64) float64
	Confparse(configuration string, r *http.Request, queryData interface{}) (result interface{})
	Dogparse(p string, r *http.Request, queryData interface{}, values map[string]interface{}) (result string)
	Attr(name, element string, data interface{}) (result interface{})
	UUID() string
	Rand() string
	Hash(str string) string
	Timenow() time.Time
	Timeformat(str interface{}, mask, format string) string
	Timetostring(time time.Time, format string) string
	Timeyear(t time.Time) string
	Timemount(t time.Time) string
	Timeday(t time.Time) string
	Timeparse(str, mask string) (res time.Time, err error)
	Refind(mask, str string, n int) (res [][]string)
	Rereplace(str, mask, new string) (res string)
	Parseparam(str string, configuration, data interface{}, resulttype string) (result interface{})
	Timefresh(str interface{}) string
	TimeExpired(str interface{}) bool
	Invert(str string) string
	Join(slice []string, sep string) (result string)
	Split(str, sep string) (result interface{})
	Tomoney(str, dec string) (res string)
	Contains1(message, str, substr string) string
	Contains(str, substr, message, messageelse string) string
	Datetotext(str string) (result string)
	Replace(str, old, new string, n int) (message string)
	Compare(var1, var2, message string) string
	Dict(values ...interface{}) (map[string]interface{}, error)
	Set(d map[string]interface{}, key string, value interface{}) map[string]interface{}
	Get(d map[string]interface{}, key string) (value interface{})
	Deletekey(d map[string]interface{}, key string) (value string)
	Sum(res, i int) int
	Cut(res string, i int, sep string) string
	Substring(str string, args ...int) string
	AddFloat(i ...interface{}) float64
	Tostring(i interface{}) (res string)
	Totree(i interface{}, objstart string) (res interface{})
	Tohtml(i interface{}) template.HTML
	Toint(i interface{}) (res int)
	Tofloat(i interface{}) (res float64)
	Tointerface(input interface{}) (res interface{})
	Concatination(values ...string) (res string)
	Marshal(i interface{}) (res string)
	Unmarshal(i string) (res interface{})
	Value(element string, configuration, data interface{}) (result interface{})
	Output(element string, configuration, data interface{}, resulttype string) (result interface{})
	ObjFromID(dt []models.Data, id string) (result interface{})
	JsonEscape(i string) (result string)

	RequestToInRequest(r *http.Request) (result model.ServiceIn)
}

// GetFuncMap возвращаем значение карты функции
func (t *tplfunc) GetFuncMap() template.FuncMap {

	funcMap := applib.FuncMap

	if len(funcMap) == 0 {
		logger.Error(context.Background(), "empty FuncMap", zap.String("place", "GetFuncMap"))
	}

	// добавляем карту функций FuncMap функциями из библиотеки github.com/Masterminds/sprig
	// только те, которые не описаны в FuncMap самостоятельно
	for k, v := range FuncMapS {
		if _, found := funcMap[k]; !found {
			funcMap[k] = v
		}
	}

	return funcMap
}

// отдаем значение куки
func (t *tplfunc) Cookie(name string, field string, r *http.Request) (result string) {
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

// получаем значение из переданного объекта
func (t *tplfunc) AddFloat(i ...interface{}) (result float64) {
	for _, b := range i {
		result += t.Tofloat(b)
	}
	return result
}

// Separator формируем сепаратор для текущей ОС
func (t *tplfunc) Separator() string {
	fm := t.GetFuncMap()
	template.New("name").Funcs(fm).ParseFiles("tplName")
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
func (t *tplfunc) Sendmail(server, port, user, pass, from, to, subject, message, turbo string) (result string) {
	var resMessage interface{}
	var fromFull, toFull, subjectFull string
	result = "true"

	f := func() {

		auth := smtp.PlainAuth("", user, pass, server)

		// приводим к одному виду чтобы можно было использвоать и <> и []
		from = t.Replace(from, ">", "]", -1)
		from = t.Replace(from, "<", "[", -1)
		to = t.Replace(to, ">", "]", -1)
		to = t.Replace(to, "<", "[", -1)

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

		from = t.Replace(from, "]", ">", -1)
		from = t.Replace(from, "[", "<", -1)
		to = t.Replace(to, "]", ">", -1)
		to = t.Replace(to, "[", "<", -1)
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
		if err := smtp.SendMail(server+":"+port, auth, t.Join(addrFrom, ","), addrTo, []byte(sendMes)); err != nil {
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

// экранируем "
// fmt.Println(jsonEscape(`dog "fish" cat`))
// output: dog \"fish\" cat
func (u *tplfunc) JsonEscape(i string) (result string) {
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	s := string(b)

	return s[1 : len(s)-1]
}

// экранируем " кроме аперсанда (&)
func (u *tplfunc) JsonEscapeUnlessAmp(s string) (result string) {
	result = u.JsonEscape(s)
	result = strings.Replace(result, `\u0026`, "&", -1)
	return result
}

func (t *tplfunc) Divfloat(a, b interface{}) float64 {
	aF := fmt.Sprint(a)
	bF := fmt.Sprint(b)
	fa, err := strconv.ParseFloat(aF, 64)
	fb, err := strconv.ParseFloat(bF, 64)

	if err != nil {
		return 0
	}

	return fa / fb
}

// умножение с запятой
func (t *tplfunc) Mulfloat(a float64, v ...float64) float64 {
	for _, b := range v {
		a = a * b
	}

	return a
}

// обработка @-функций внутри конфигурации (в шаблонизаторе)
func (t *tplfunc) Confparse(configuration string, r *http.Request, queryData interface{}) (result interface{}) {
	var d models.Data
	var frml = New(t.cfg, t.api)

	b, err := json.Marshal(queryData)
	json.Unmarshal(b, &d)

	if err != nil {
		return "Error! Failed marshal queryData: " + fmt.Sprint(err)
	}
	dv := []models.Data{d}
	confParse, _ := frml.Exec(configuration, dv, nil, t.RequestToInRequest(r), "")

	// конфигурация с обработкой @-функции
	var conf map[string]model.Element
	if confParse != "" {
		err = json.Unmarshal([]byte(confParse), &conf)
	}

	if err != nil {
		return "Error! Failed unmarshal confParse: " + fmt.Sprint(err) + " - " + confParse
	}

	return conf
}

// обработка @-функций внутри шаблонизатора
func (t *tplfunc) Dogparse(p string, r *http.Request, queryData interface{}, values map[string]interface{}) (result string) {
	var frml = New(t.cfg, t.api)
	var d models.Data

	b, _ := json.Marshal(queryData)
	json.Unmarshal(b, &d)

	dv := []models.Data{d}
	result, _ = frml.Exec(p, dv, values, t.RequestToInRequest(r), "")

	return result
}

// получаем значение из переданного объекта
func (t *tplfunc) Attr(name, element string, data interface{}) (result interface{}) {
	var dt models.Data
	json.Unmarshal([]byte(t.Marshal(data)), &dt)

	dtl := &dt
	result, _ = dtl.Attr(name, element)

	return result
}

func (t *tplfunc) UUID() string {
	stUUID := uuid.NewV4()
	return stUUID.String()
}

func (t *tplfunc) Rand() string {
	uuid := t.UUID()
	return uuid[1:6]
}

func (t *tplfunc) Hash(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	return sha1_hash
}

func (t *tplfunc) Timenow() time.Time {
	return time.Now().UTC()
}

// преобразуем текст в дату (если ошибка - возвращаем false), а потом обратно в строку нужного формата
func (t *tplfunc) Timeformat(str interface{}, mask, format string) string {
	ss := fmt.Sprintf("%v", str)

	timeFormat, err := t.Timeparse(ss, mask)
	if err != nil {
		return fmt.Sprint(err)
	}
	res := timeFormat.Format(format)

	return res
}

// преобразуем текст в дату (если ошибка - возвращаем false), а потом обратно в строку нужного формата
func (t *tplfunc) Timetostring(time time.Time, format string) string {
	res := time.Format(format)

	return res
}

func (t *tplfunc) Timeyear(tm time.Time) string {
	ss := fmt.Sprintf("%v", tm.Year())
	return ss
}

func (t *tplfunc) Timemount(tm time.Time) string {
	ss := fmt.Sprintf("%v", tm.Month())
	return ss
}

func (t *tplfunc) Timeday(tm time.Time) string {
	ss := fmt.Sprintf("%v", tm.Day())
	return ss
}

func (t *tplfunc) Timeparse(str, mask string) (res time.Time, err error) {
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

// альтернативный формат добавления даты год-месяц-день (0-1-0)
//func (t *tplfunc) timeaddday(t time.Time, dateformat string) (input time.Time, error string) {
//	var intervalYMD []int
//	intervalSl := strings.Split(dateformat, "-")
//
//	if len(intervalSl) != 3 {
//		return input, "Error! Params failed. (want: year-mount-day, e.g. 0-0-1)"
//	}
//
//	i0, err := strconv.Atoi(intervalSl[0])
//	if err != nil {
//		return input, fmt.Sprintln(err)
//	}
//	intervalYMD = append(intervalYMD, i0)
//
//	i1, err := strconv.Atoi(intervalSl[1])
//	if err != nil {
//		return input, fmt.Sprintln(err)
//	}
//	intervalYMD = append(intervalYMD, i1)
//
//	i2, err := strconv.Atoi(intervalSl[2])
//	if err != nil {
//		return input, fmt.Sprintln(err)
//	}
//	intervalYMD = append(intervalYMD, i2)
//
//	input = t.AddDate(intervalYMD[0], intervalYMD[1], intervalYMD[2])
//
//	return input, ""
//}

func (t *tplfunc) Refind(mask, str string, n int) (res [][]string) {
	if n == 0 {
		n = -1
	}
	re := regexp.MustCompile(mask)
	res = re.FindAllStringSubmatch(str, n)

	return
}

func (t *tplfunc) Rereplace(str, mask, new string) (res string) {
	re := regexp.MustCompile(mask)
	res = re.ReplaceAllString(str, new)

	return
}

func (t *tplfunc) Parseparam(str string, configuration, data interface{}, resulttype string) (result interface{}) {

	// разбиваем строку на слайс для замкены и склейки
	sl := strings.Split(str, "%")

	if len(sl) > 1 {
		for k, v := range sl {
			// нечетные значения - это выделенные переменные
			if k == 1 || k == 3 || k == 5 || k == 7 || k == 9 || k == 11 || k == 13 || k == 15 || k == 17 || k == 19 || k == 21 || k == 23 {
				res := t.Output(v, configuration, data, resulttype)
				sl[k] = fmt.Sprint(res)
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
func (t *tplfunc) Timefresh(str interface{}) string {

	ss := fmt.Sprintf("%v", str)
	start := time.Now().UTC()

	format := "2006-01-02 15:04:05 +0000 UTC"
	end, _ := time.Parse(format, ss)

	if start.After(end) {
		return "true"
	}

	return "false"
}

// функция указывает что указанное время истекло
// относительно текущего времени
func (t *tplfunc) TimeExpired(str interface{}) bool {

	ss := fmt.Sprintf("%v", str)
	start := time.Now().UTC()

	timeFormat := "2006-01-02 15:04:05 +0000 UTC"
	end, _ := time.Parse(timeFormat, ss)

	if start.After(end) {
		return true
	}

	return false
}

// инвертируем строку
func (t *tplfunc) Invert(str string) string {
	var result string
	for i := len(str); i > 0; i-- {
		result = result + string(str[i-1])
	}
	return result
}

// переводим массив в строку
func (t *tplfunc) Join(slice []string, sep string) (result string) {
	result = strings.Join(slice, sep)

	return result
}

// разбиваем строку на массив
func (t *tplfunc) Split(str, sep string) (result interface{}) {
	result = strings.Split(str, sep)

	return result
}

// переводим в денежное отображение строки - 12.344.342
func (t *tplfunc) Tomoney(str, dec string) (res string) {

	for i, v1 := range t.Invert(str) {
		if (i == 3) || (i == 6) || (i == 9) {
			if (string(v1) != " ") && (string(v1) != "+") && (string(v1) != "-") {
				res = res + dec
			}
		}
		res = res + string(v1)
	}
	return t.Invert(res)
}

func (t *tplfunc) Contains1(message, str, substr string) string {
	sl1 := strings.Split(substr, "|")
	for _, v := range sl1 {
		if strings.Contains(str, v) {
			return message
		}
	}
	return ""
}

func (t *tplfunc) Contains(str, substr, message, messageelse string) string {
	sl1 := strings.Split(substr, "|")
	for _, v := range sl1 {
		if strings.Contains(str, v) {
			return message
		}
	}
	return messageelse
}

// преобразую дату из 2013-12-24 в 24 января 2013
func (t *tplfunc) Datetotext(str string) (result string) {
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
func (t *tplfunc) Replace(str, old, new string, n int) (message string) {
	message = strings.Replace(str, old, new, n)
	return message
}

// сравнивает два значения и вы	водит текст, если они равны
func (t *tplfunc) Compare(var1, var2, message string) string {
	if var1 == var2 {
		return message
	}
	return ""
}

// фукнцкия мультитпликсирования передаваемых параметров при передаче в шаблонах нескольких параметров
// {{template "sub-template" dict "Data" . "Values" $.Values}}
func (t *tplfunc) Dict(values ...interface{}) (map[string]interface{}, error) {
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

func (t *tplfunc) Set(d map[string]interface{}, key string, value interface{}) map[string]interface{} {
	d[key] = value
	return d
}

func (t *tplfunc) Get(d map[string]interface{}, key string) (value interface{}) {
	value, found := d[key]
	if !found {
		value = ""
	}
	return value
}

func (t *tplfunc) Deletekey(d map[string]interface{}, key string) (value string) {
	delete(d, key)
	return "true"
}

// суммируем
func (t *tplfunc) Sum(res, i int) int {
	res = res + i
	return res
}

// образаем по заданному кол-ву символов
func (t *tplfunc) Cut(res string, i int, sep string) string {
	res = strings.Trim(res, " ")
	if i <= len([]rune(res)) {
		res = string([]rune(res)[:i]) + sep
	}
	return res
}

// обрезаем строку (строка, откуда, [сколько])
// если откуда отрицательно, то обрезаем с конца
func (t *tplfunc) Substring(str string, args ...int) string {
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

func (t *tplfunc) Tostring(i interface{}) (res string) {
	res = fmt.Sprint(i)
	return res
}

// функция преобразует переданные данные в формат типа Items с вложенными подпукнтами
func (t *tplfunc) Totree(i interface{}, objstart string) (res interface{}) {
	var objD []models.Data
	var objRes []interface{}
	var objTree []models.DataTreeOut

	b3, _ := json.Marshal(i)
	err := json.Unmarshal(b3, &objD)
	if err != nil {
		return "Error convert to ResponseData. err: " + fmt.Sprint(err)
	}

	in := t.tree.DataToIncl(objD)
	resTree := t.tree.TreeShowIncl(in, objstart)

	// наполняем дерево
	for _, v := range resTree {
		objRes = append(objRes, *v)
	}

	b, _ := json.Marshal(objRes)
	err = json.Unmarshal(b, &objTree)
	if err != nil {
		return "Error convert to DataTree"
	}

	return objTree
}

func (t *tplfunc) Tohtml(i interface{}) template.HTML {

	return template.HTML(i.(string))
}

func (t *tplfunc) Toint(i interface{}) (res int) {
	str := fmt.Sprint(i)
	i = strings.Trim(str, " ")
	res, err := strconv.Atoi(str)
	if err != nil {
		return -1
	}

	return res
}

func (t *tplfunc) Tofloat(i interface{}) (res float64) {
	str := fmt.Sprint(i)
	str = strings.Trim(str, " ")
	str = strings.ReplaceAll(str, ",", ".")
	res, e := strconv.ParseFloat(str, 10)
	if e != nil {
		return -1
	}

	return res
}

func (t *tplfunc) Tointerface(input interface{}) (res interface{}) {
	b3, _ := json.Marshal(input)
	err := json.Unmarshal(b3, &res)
	if err != nil {
		return err
	}
	return
}

func (t *tplfunc) Concatination(values ...string) (res string) {
	res = strings.Join(values, "")
	return res
}

func (t *tplfunc) Marshal(i interface{}) (res string) {
	b3, _ := json.Marshal(i)
	return string(b3)
}

func (t *tplfunc) Unmarshal(i string) (res interface{}) {
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
func (t *tplfunc) Value(element string, configuration, data interface{}) (result interface{}) {
	var conf map[string]model.Element
	json.Unmarshal([]byte(t.Marshal(configuration)), &conf)

	var dt model.Data
	json.Unmarshal([]byte(t.Marshal(data)), &dt)

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

				json.Unmarshal([]byte(t.Marshal(v.Source)), &source)
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

// получить объект из массива объектов по id
func (t *tplfunc) ObjFromID(dt []models.Data, id string) (result interface{}) {
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

// получаем значение из массива данных по имени элемента
// ПЕРЕДЕЛАТЬ! приходится постоянно сериализовать данные
func (t *tplfunc) Output(element string, configuration, data interface{}, resulttype string) (result interface{}) {
	var conf map[string]model.Element
	json.Unmarshal([]byte(t.Marshal(configuration)), &conf)

	if resulttype == "" {
		resulttype = "text"
	}

	var dt model.Data
	json.Unmarshal([]byte(t.Marshal(data)), &dt)

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

				json.Unmarshal([]byte(t.Marshal(v.Source)), &source)
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

			//if v.Type == "object" {
			//	var args []string
			//
			//	json.Unmarshal([]byte(marshal(v.Source)), &args)
			//	result = Obj(dt, args)
			//}

		}
	}

	if resulttype == "html" {
		result = template.HTML(result.(string))
	}

	return result
}

// RequestToInRequest вспомогательная функция только для HTTP
// преобразуем полученный через прямой параметр запрос в темплейте
// в формат model.ServiceIn в котором его понимаеют обработчики функций
func (t *tplfunc) RequestToInRequest(r *http.Request) (result model.ServiceIn) {
	vars := mux.Vars(r)
	result.Page = vars["page"]

	result.Url = r.URL.Query().Encode()
	result.Referer = r.Referer()
	result.RequestURI = r.RequestURI
	result.Form = r.Form
	result.Host = r.Host
	result.Query = r.URL.Query()

	// указатель на профиль текущего пользователя
	var profile models.ProfileData
	profileRaw := r.Context().Value("UserRaw")
	json.Unmarshal([]byte(fmt.Sprint(profileRaw)), &profile)

	result.Profile = profile

	return
}

func NewTplFunc(cfg model.Config, api api.Api) TplFunc {
	tree := tree.New(
		cfg,
	)

	r := &tplfunc{
		cfg:  cfg,
		tree: tree,
		api:  api,
	}

	return r
}
