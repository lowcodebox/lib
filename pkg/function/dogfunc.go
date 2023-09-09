package function

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	api "git.lowcodeplatform.net/fabric/api-client"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/models"
	uuid "github.com/satori/go.uuid"
)

type function struct {
	cfg     model.Config
	formula Formula
	dogfunc DogFunc
	tplfunc TplFunc
	api     api.Api
}

type Function interface {
	Exec(p string, queryData []models.Data, values map[string]interface{}, request model.ServiceIn, blockname string) (result string, err error)
	TplFunc() TplFunc
}

////////////////////////////////////////////////////////////

type formula struct {
	value    string        `json:"value"`
	document []models.Data `json:"document"`
	request  model.ServiceIn
	inserts  []*insert
	values   map[string]interface{} //  параметры переданные в шаблон при генерации страницы (доступны в шаблоне как $.Value)
	cfg      model.Config
	dogfunc  DogFunc
}

type Formula interface {
	Replace() (result string, err error)
	Parse() (err error)
	Calculate() (err error)

	SetValue(value string)
	SetValues(value map[string]interface{})
	SetDocument(value []models.Data)
	SetRequest(value model.ServiceIn)
	SetInserts(value []*insert)
}

// Вставка - это одна функция, которая может иметь вложения
// Text - строка вставки, по которому мы будем заменять в общем тексте
type insert struct {
	text      string      `json:"text"`
	arguments []string    `json:"arguments"`
	result    interface{} `json:"result"`
	dogfuncs  dogfunc
}

// Исчисляемая фукнция с аргументами и параметрами
// может иметь вложения
type dogfunc struct {
	name      string       `json:"name"`
	arguments []string     `json:"arguments"`
	result    interface{}  `json:"result"`
	cfg       model.Config `json:"cfg"`
	tplfunc   TplFunc
	api       api.Api
}

type DogFunc interface {
	Query(r *http.Request, arg []string) (result interface{}, err error)
	TplValue(v map[string]interface{}, arg []string) (result string, err error)
	ConfigValue(arg []string) (result string, err error)
	SplitIndex(arg []string) (result string, err error)
	Time(arg []string) (result string, err error)
	TimeFormat(arg []string) (result string, err error)
	FuncURL(r model.ServiceIn, arg []string) (result string, err error)
	Cookie(r model.ServiceIn, arg []string) (result string, err error)
	Path(d []models.Data, arg []string) (result string, err error)
	DReplace(arg []string) (result string, err error)
	UserObj(r model.ServiceIn, arg []string) (result string, err error)
	UserProfile(r model.ServiceIn, arg []string) (result string, err error)
	UserRole(r model.ServiceIn, arg []string) (result string, err error)
	Obj(data []models.Data, arg []string) (result string, err error)
	FieldValue(data []models.Data, arg []string) (result string, err error)
	FieldSrc(data []models.Data, arg []string) (result string, err error)
	FieldSplit(data []models.Data, arg []string) (result string, err error)
	DateModify(arg []string) (result string, err error)
	DogSendmail(arg []string) (result string, err error)
}

////////////////////////////////////////////////////////////
// !!! ПОКА ТОЛЬКО ПОСЛЕДОВАТЕЛЬНАЯ ОБРАБОТКА (без сложений)
////////////////////////////////////////////////////////////

func (p *formula) Replace() (result string, err error) {
	err = p.Parse()
	if err != nil {
		return "", err
	}

	err = p.Calculate()
	if err != nil {
		return "", err
	}

	for _, v := range p.inserts {
		p.value = strings.Replace(p.value, v.text, fmt.Sprint(v.result), -1)
	}

	return p.value, err
}

func (p *formula) Parse() (err error) {

	// пропускаем если пришло пустое значение. ошибки нет, просто параметр не передали
	if p.value == "" {
		return nil
		//return fmt.Errorf("%s", "[Parse] Error. Value is empty.")
	}

	value := p.value

	pattern := regexp.MustCompile(`@(\w+)\(\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?\)`)
	allIndexes := pattern.FindAllStringSubmatch(value, -1)

	for _, loc := range allIndexes {

		i := insert{}
		f := dogfunc{}
		i.dogfuncs = f
		p.inserts = append(p.inserts, &i)

		strFunc := string(loc[0])

		strFunc1 := strings.Replace(strFunc, "@", "", -1)
		strFunc1 = strings.Replace(strFunc1, ")", "", -1)
		f1 := strings.Split(strFunc1, "(")

		if len(f1) == 1 { // если не нашли ( значит неверно задана @-фукнций
			return fmt.Errorf("%s", "[Parse] Error. Len strFunc1 value is 1.")
		}

		i.text = strFunc
		i.dogfuncs.name = f1[0] // название функции

		// готовим параметры для передачи в функцию обработки
		if len(f1[1]) > 0 {
			arg := f1[1]

			// разбиваем по запятой
			args := strings.Split(arg, ",")

			// очищаем каждый параметр от ' если есть
			argsClear := []string{}
			for _, v := range args {
				v = strings.Trim(v, " ")
				v = strings.Trim(v, "'")
				argsClear = append(argsClear, v)
			}
			i.dogfuncs.arguments = argsClear
		}

		//for j, loc1 := range loc {
		//	res := string(loc1)
		//	if res == "" {
		//		continue
		//	}
		//
		//	// общий текст вставки
		//	if j == 0 {
		//		i.Text = res
		//	}
		//
		//	// название фукнции
		//	if j == 1 {
		//		i.dogfuncs.Name = res
		//	}
		//
		//	// аргументы для функции
		//	if j > 1 {
		//		if strings.Contains(res, "@") {
		//			// рекурсивно парсим вложенную формулу
		//			//argRec := p.Parse(arg)
		//			//f.Arguments = append(f.Arguments, argRec)
		//		} else {
		//			// добавляем аргумент в слайс аргументов
		//			i.dogfuncs.Arguments = append(i.dogfuncs.Arguments, res)
		//		}
		//	}
		//}
	}

	return err
}

func (p *formula) Calculate() (err error) {
	var result interface{}

	for k, v := range p.inserts {
		param := strings.ToUpper(v.dogfuncs.name)

		switch param {
		case "QUERY":
			result, err = p.dogfunc.Query(p.request.RequestRaw, v.dogfuncs.arguments)
		case "RAND":
			uuid := uuid.NewV4().String()
			result = uuid[1:6]
		case "SENDMAIL":
			result, err = p.dogfunc.DogSendmail(v.dogfuncs.arguments)
		case "PATH":
			result, err = p.dogfunc.Path(p.document, v.dogfuncs.arguments)
		case "REPLACE":
			result, err = p.dogfunc.DReplace(v.dogfuncs.arguments)
		case "TIME":
			result, err = p.dogfunc.Time(v.dogfuncs.arguments)
		case "DATEMODIFY":
			result, err = p.dogfunc.DateModify(v.dogfuncs.arguments)

		case "USER":
			result, err = p.dogfunc.UserObj(p.request, v.dogfuncs.arguments)
		case "ROLE":
			result, err = p.dogfunc.UserRole(p.request, v.dogfuncs.arguments)
		case "PROFILE":
			result, err = p.dogfunc.UserProfile(p.request, v.dogfuncs.arguments)

		case "OBJ":
			result, err = p.dogfunc.Obj(p.document, v.dogfuncs.arguments)
		case "URL":
			result, err = p.dogfunc.FuncURL(p.request, v.dogfuncs.arguments)
		case "COOKIE":
			result, err = p.dogfunc.Cookie(p.request, v.dogfuncs.arguments)

		case "SPLITINDEX":
			result, err = p.dogfunc.SplitIndex(v.dogfuncs.arguments)

		case "TPLVALUE":
			result, err = p.dogfunc.TplValue(p.values, v.dogfuncs.arguments)
		case "CONFIGVALUE":
			result, err = p.dogfunc.ConfigValue(v.dogfuncs.arguments)
		case "FIELDVALUE":
			result, err = p.dogfunc.FieldValue(p.document, v.dogfuncs.arguments)
		case "FIELDSRC":
			result, err = p.dogfunc.FieldSrc(p.document, v.dogfuncs.arguments)
		case "FIELDSPLIT":
			result, err = p.dogfunc.FieldSplit(p.document, v.dogfuncs.arguments)
		default:
			result = ""
		}

		if err != nil {
			p.inserts[k].result = fmt.Sprint(err)
		} else {
			p.inserts[k].result = result
		}
	}

	return err
}

func (f *formula) SetValue(value string) {
	f.value = value
}

func (f *formula) SetValues(value map[string]interface{}) {
	f.values = value
}

func (f *formula) SetDocument(value []models.Data) {
	f.document = value
}

func (f *formula) SetRequest(value model.ServiceIn) {
	f.request = value
}

func (f *formula) SetInserts(value []*insert) {
	f.inserts = value
}

func NewFormula(cfg model.Config, dogfunc DogFunc) Formula {
	return &formula{
		cfg:     cfg,
		dogfunc: dogfunc,
	}
}

///////////////////////////////////////////////////
// Фукнции @ обработки
///////////////////////////////////////////////////

// Делаем вложенный запрос
// аргументы:
// queryName - первый параметр - имя запрсоа;
// mode - тип ответа
// 		id (по-умолчанию) 	- список UID-ов
// 		data 				- []Data
//		response			- полный ответ формате Response
func (d *dogfunc) Query(r *http.Request, arg []string) (result interface{}, err error) {
	var objs models.Response
	var mode = "id"
	if len(arg) == 0 {
		return nil, fmt.Errorf("%s", "Ошибка в переданных параметрах.")
	}
	if len(arg) == 2 {
		mode = arg[1]
	}

	formValues := r.PostForm
	bodyJSON, _ := json.Marshal(formValues)

	// добавляем к пути в запросе переданные в блок параметры ULR-а (возможно там есть параметры для фильтров)
	filters := r.URL.RawQuery
	if filters != "" {
		filters = "?" + filters
	}

	res, _ := d.api.Query(arg[0]+filters, "GET", string(bodyJSON))
	json.Unmarshal([]byte(res), &objs)
	//d.utl.Curl("GET", "query/"+arg[0]+filters, string(bodyJSON), &objs, map[string]string{})

	switch mode {
	case "data":
		return objs.Data, err
	case "response":
		return objs, err
	default:
		var resUIDs []string
		var respData []models.Data

		// если можно привести, значит формат внутреннего запроса и возвращаем список uid
		r, _ := json.Marshal(objs.Data)
		err := json.Unmarshal(r, &respData)
		if err != nil {
			return "Error execute dogfunc Query", err
		}
		for _, v := range respData {
			resUIDs = append(resUIDs, v.Uid)
		}
		return strings.Join(resUIDs, ","), err
	}

	return "", err
}

// Получение значений $.Value шаблона (работает со значением по-умолчанию)
func (d *dogfunc) TplValue(v map[string]interface{}, arg []string) (result string, err error) {
	var valueDefault string

	if len(arg) > 0 {
		param := arg[0]
		if len(arg) == 2 {
			valueDefault = arg[1]
		}

		result, found := v[strings.Trim(param, " ")]
		if !found {
			if valueDefault == "" {
				return "", fmt.Errorf("%s", "Error parsing @-dogfunc TplValue (Value from this key is not found.)")
			}
			return fmt.Sprint(valueDefault), err
		}
		return fmt.Sprint(result), err

	} else {
		return "", fmt.Errorf("%s", "Error parsing @-dogfunc TplValue (Arguments is null)")
	}

	return fmt.Sprint(result), err
}

// Получение значений из конфигурации проекта (хранится State в объекте приложение App)
func (d *dogfunc) ConfigValue(arg []string) (result string, err error) {
	var valueDefault string

	if len(arg) > 0 {
		param := arg[0]
		if len(arg) == 2 {
			valueDefault = arg[1]
		}

		result, err := d.cfg.GetValue(strings.Trim(param, " "))
		if err != nil {
			if valueDefault == "" {
				return "", fmt.Errorf("%s", "Error parsing @-dogfunc ConfigValue (Value from this key is not found.)")
			}
			return fmt.Sprint(valueDefault), err
		}
		return fmt.Sprint(result), err

	} else {
		return "", fmt.Errorf("%s", "Error parsing @-dogfunc ConfigValue (Arguments is null)")
	}

	return fmt.Sprint(result), err
}

// Получаем значение из разделенной строки по номер
// параметры:
// str - текст (строка)
// sep - разделитель (строка)
// index - порядковый номер (число) (от 0) возвращаемого элемента
// default - значение по-умолчанию (не обязательно)
func (d *dogfunc) SplitIndex(arg []string) (result string, err error) {
	var valueDefault string

	if len(arg) > 0 {

		str := d.tplfunc.Replace(arg[0], "'", "", -1)
		sep := d.tplfunc.Replace(arg[1], "'", "", -1)
		index := d.tplfunc.Replace(arg[2], "'", "", -1)
		defaultV := d.tplfunc.Replace(arg[3], "'", "", -1)

		in, err := strconv.Atoi(index)
		if err != nil {
			return result, err
		}
		if len(arg) == 4 {
			valueDefault = defaultV
		}

		slice_str := strings.Split(str, sep)
		result = slice_str[in]
	}
	if result == "" {
		result = valueDefault
	}

	return result, err
}

// Получение текущей даты
func (d *dogfunc) Time(arg []string) (result string, err error) {

	if len(arg) > 0 {
		param := strings.ToUpper(arg[0])

		switch param {
		case "NOW", "THIS":
			result = time.Now().Format("2006-01-02 15:04:05")
		default:
			result = time.Now().String()
		}
	}

	return
}

// Получение идентификатор User-а
func (d *dogfunc) TimeFormat(arg []string) (result string, err error) {
	var valueDefault string

	if len(arg) > 0 {

		thisdate := strings.ToUpper(arg[0]) // переданное время (строка) можно вручную или Now (текущее)
		mask := strings.ToUpper(arg[1])     // маска для перевода переданного времени в Time
		format := strings.ToUpper(arg[2])   // формат преобразования времени (как вывести)
		if len(arg) == 4 {
			valueDefault = strings.ToUpper(arg[2])
		}

		ss := thisdate
		switch thisdate {
		case "NOW":
			ss = time.Now().UTC().String()
			mask = "2006-01-02 15:04:05"
		}

		result = d.tplfunc.Timeformat(ss, mask, format)
	}
	if result == "" {
		result = valueDefault
	}

	return
}

func (d *dogfunc) FuncURL(r model.ServiceIn, arg []string) (result string, err error) {
	var valueDefault string
	if len(arg) > 0 {
		param := arg[0]
		result = strings.Join(r.Form[param], ",")

		if len(arg) == 2 {
			valueDefault = arg[1]
		}
	}
	if result == "" {
		result = valueDefault
	}

	return
}

// Получение cookie
// параметры: 	1-й параметр - name (имя куки)
//				2-й параметр - field (поле куки: VALUE, EXPIRES, MAXAGE, PATH, SECURE)
func (d *dogfunc) Cookie(r model.ServiceIn, arg []string) (result string, err error) {
	if len(arg) > 0 {

		name := arg[0]
		field := strings.ToUpper(arg[1])

		c, err := r.RequestRaw.Cookie(name)
		if err != nil {
			return "", err
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
	}
	return result, err
}

// Вставляем значения системных полей объекта
func (d *dogfunc) Path(dm []models.Data, arg []string) (result string, err error) {
	var valueDefault string

	if len(arg) > 0 {
		param := strings.ToUpper(arg[0])

		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

		switch param {
		case "API":
			result = d.cfg.UrlApi
		case "GUI":
			result = d.cfg.UrlGui
		case "PROXY":
			result = d.cfg.UrlProxy
		case "CLIENT":
			result = d.cfg.ClientPath
		case "DOMAIN":
			result = d.cfg.Domain
		default:
			result = d.cfg.ClientPath
		}
	}

	if result == "" {
		result = valueDefault
	}

	return
}

// Заменяем значение
func (d *dogfunc) DReplace(arg []string) (result string, err error) {
	var count int
	var str, oldS, newS string

	if len(arg) > 0 {
		str = arg[0]
		oldS = arg[1]
		newS = arg[2]

		if len(arg) >= 4 {
			countString := arg[3]
			count, err = strconv.Atoi(countString)
			if err != nil {
				count = -1
			}
		}
		result = strings.Replace(str, oldS, newS, count)
	}

	return
}

// Получение идентификатор User-а (для Cockpit-a)
func (d *dogfunc) UserObj(r model.ServiceIn, arg []string) (result string, err error) {
	var valueDefault string

	if len(arg) > 0 {

		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

		param := strings.ToUpper(arg[0])
		uu := r.Profile // текущий профиль пользователя

		if &uu != nil {
			switch param {
			case "UID", "ID":
				result = uu.Uid
			case "PHOTO":
				result = uu.Photo
			case "AGE":
				result = uu.Age
			case "NAME":
				result = uu.FirstName + " " + uu.LastName
			case "EMAIL":
				result = uu.Email
			case "STATUS":
				result = uu.Status
			default:
				result = uu.Uid
			}
		}

	}
	if result == "" {
		result = valueDefault
	}

	return
}

// Получение UserProfile (для Cockpit-a)
func (d *dogfunc) UserProfile(r model.ServiceIn, arg []string) (result string, err error) {
	if len(arg) > 0 {

		param := strings.ToUpper(arg[0])

		var uu = r.Profile
		role := uu.CurrentRole

		switch param {
		case "UID", "ID":
			result = role.Uid
		case "TITLE":
			result = role.Title
		case "DEFAULT":
			result, _ = role.Attr("profile_default", "value")
		default:
			result = uu.Uid
		}

	}
	return
}

// Получение текущей роли User-а
func (d *dogfunc) UserRole(r model.ServiceIn, arg []string) (result string, err error) {
	if len(arg) > 0 {

		param := strings.ToUpper(arg[0])
		param2 := strings.ToUpper(arg[1])

		var uu = r.Profile
		role := uu.CurrentRole

		switch param {
		case "UID", "ID":
			result = role.Uid
		case "TITLE":
			result = role.Title
		case "ADMIN":
			result, _ = role.Attr("role_default", "value")
		case "HOMEPAGE":
			if param2 == "SRC" {
				result, _ = role.Attr("homepage", "src")
			} else {
				result, _ = role.Attr("homepage", "value")
			}
		case "DEFAULT":
			result, _ = role.Attr("default", "value")
		default:
			result = uu.Uid
		}

	}
	return
}

// Вставляем значения системных полей объекта
func (d *dogfunc) Obj(data []models.Data, arg []string) (result string, err error) {
	var valueDefault, r string
	var res = []string{}
	if len(arg) == 0 {
		err = fmt.Errorf("%s", "Ошибка в переданных параметрах")
		return
	}

	param := strings.ToUpper(arg[0])
	separator := "," // значение разделителя по-умолчанию

	if len(arg) == 0 {
		err = fmt.Errorf("%s", "Ошибка в переданных параметрах.")
		return
	}
	if len(arg) == 2 {
		valueDefault = arg[1]
	}
	if len(arg) == 3 {
		separator = arg[2]
	}

	for _, d := range data {
		switch param {
		case "UID": // получаем все uid-ы из переданного массива объектов
			r = d.Uid
		case "ID":
			r = d.Id
		case "SOURCE":
			r = d.Source
		case "TITLE":
			r = d.Title
		case "TYPE":
			r = d.Type
		default:
			r = d.Uid
		}
		res = append(res, r)
	}
	result = d.tplfunc.Join(res, separator)

	if result == "" {
		result = valueDefault
	}

	return
}

// Вставляем значения (Value) элементов из формы
// Если поля нет, то выводит переданное значение (может быть любой символ)
func (d *dogfunc) FieldValue(data []models.Data, arg []string) (result string, err error) {
	var valueDefault, separator string
	var resSlice = []string{}

	separator = "," // значение разделителя по-умолчанию

	if len(arg) == 0 {
		err = fmt.Errorf("%s", "Ошибка в переданных параметрах.")
		return
	}

	param := arg[0]
	if len(arg) == 2 {
		valueDefault = arg[1]
	}
	if len(arg) == 3 {
		separator = arg[2]
	}

	for _, d := range data {
		val, found := d.Attr(param, "value")
		if found {
			resSlice = append(resSlice, strings.Trim(val, " "))
		}
	}
	result = d.tplfunc.Join(resSlice, separator)

	if result == "" {
		result = valueDefault
	}

	return
}

// Вставляем ID-объекта (SRC) элементов из формы
// Если поля нет, то выводит переданное значение (может быть любой символ)
func (d *dogfunc) FieldSrc(data []models.Data, arg []string) (result string, err error) {
	var valueDefault, separator string
	var resSlice = []string{}

	if len(arg) == 0 {
		err = fmt.Errorf("%s", "Ошибка в переданных параметрах.")
		return
	}

	param := arg[0]
	if len(arg) == 2 {
		valueDefault = arg[1]
	}
	if len(arg) == 3 {
		separator = arg[2]
	}

	for _, d := range data {
		val, found := d.Attr(param, "src")
		if found {
			resSlice = append(resSlice, strings.Trim(val, " "))
		}
	}
	result = d.tplfunc.Join(resSlice, separator)

	if result == "" {
		result = valueDefault
	}

	return
}

// Разбиваем значения по элементу (Value(по-умолчанию)/Src) элементов из формы по разделителю и возвращаем
// значение по указанному номеру (начала от 0)
// Синтаксис: FieldValueSplit(поле, элемент, разделитель, номер_элемента)
// для разделителя есть кодовые слова slash - / (нельзя вставить в фукнцию)
func (d *dogfunc) FieldSplit(data []models.Data, arg []string) (result string, err error) {
	var resSlice = []string{}
	var r string

	if len(arg) == 0 {
		err = fmt.Errorf("%s", "Ошибка в переданных параметрах.")
		return
	}

	if len(arg) < 4 {
		err = fmt.Errorf("%s", "Error! Count params must have 4 (field, element, separator, number)")
		return
	}
	field := arg[0]
	element := arg[1]
	sep := arg[2]
	num_str := arg[3]

	if element == "" {
		element = "value"
	}

	// 1. преобразовали в номер
	num, err := strconv.Atoi(num_str)
	if err != nil {
		return
	}

	for _, d := range data {
		// 2. получили значение поля
		val, found := d.Attr(field, element)

		if !found {
			err = fmt.Errorf("%s", "Error! This field is not found.")
			return
		}
		in := strings.Trim(val, " ")
		if sep == "slash" {
			sep = "/"
		}

		// 3. разделили и получили нужный элемент
		split_v := strings.Split(in, sep)
		if len(split_v) < num {
			err = fmt.Errorf("%s", "Error! Array size is less than the passed number")
			return
		}

		r = split_v[num]
		resSlice = append(resSlice, r)
	}

	result = d.tplfunc.Join(resSlice, ",")

	return
}

// Добавление даты к переданной
// date - дата, которую модифицируют (значение должно быть в формате времени)
// modificator - модификатор (например "+24h")
// format - формат переданного времени (по-умолчанию - 2006-01-02T15:04:05Z07:00 (формат: time.RFC3339)
func (d *dogfunc) DateModify(arg []string) (result string, err error) {

	if len(arg) < 2 {
		err = fmt.Errorf("%s", "Error! Count params must have min 2 (date, modificator; option: format)")
		return
	}
	dateArg := arg[0]
	modificator := arg[1]

	format := "2006-01-02 15:04:05"
	if len(arg) == 3 {
		format = arg[2]
	}

	// преобразуем полученную дату из строки в дату
	date, err := time.Parse(format, dateArg)
	if err != nil {
		return
	}

	// преобразуем модификатор во время
	p, err := time.ParseDuration(modificator)
	if err != nil {
		return
	}

	return fmt.Sprint(date.Add(p)), err
}

///////////////////////////////////////////////////////////////
// Отправляем почтового сообщения
func (d *dogfunc) DogSendmail(arg []string) (result string, err error) {
	if len(arg) < 9 {
		err = fmt.Errorf("%s", "Error! Count params must have min 9 (server, port, user, pass, from, to, subject, message, turbo: string)")
		return
	}
	result = d.tplfunc.Sendmail(arg[0], arg[1], arg[2], arg[3], arg[4], arg[5], arg[6], arg[7], arg[8])

	return
}

func NewDogFunc(cfg model.Config, tplfunc TplFunc, api api.Api) DogFunc {
	return &dogfunc{
		cfg:     cfg,
		tplfunc: tplfunc,
		api:     api,
	}
}

///////////////////////////////////////////////////////////////
// Собачья-обработка (поиск в строке @функций и их обработка)
///////////////////////////////////////////////////////////////
func (d *function) Exec(p string, queryData []models.Data, values map[string]interface{}, request model.ServiceIn, blockname string) (result string, err error) {

	var fml = NewFormula(d.cfg, d.dogfunc)
	if p == "" {
		return
	}

	// прогоняем полученную строку такое кол-во раз, сколько вложенных уровней + 1 (для сравнения)
	for {
		fml.SetValue(p)
		fml.SetValues(values)
		fml.SetDocument(queryData)
		fml.SetRequest(request)
		res_parse, err := fml.Replace()

		if err != nil {
			return result, err
		}

		if p == res_parse {
			result = res_parse
			break
		}
		p = res_parse
	}

	return result, err
}

func (d *function) TplFunc() TplFunc {
	return d.tplfunc
}

func New(cfg model.Config, api api.Api) Function {
	tplfunc := NewTplFunc(cfg, api)
	dogfunc := NewDogFunc(cfg, tplfunc, api)
	formula := NewFormula(cfg, dogfunc)

	return &function{
		cfg:     cfg,
		formula: formula,
		dogfunc: dogfunc,
		tplfunc: tplfunc,
		api:     api,
	}
}
