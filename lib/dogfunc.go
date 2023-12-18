package app_lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.lowcodeplatform.net/packages/logger"
	"go.uber.org/zap"
)

////////////////////////////////////////////////////////////

type Formula struct {
	Value    string `json:"value"`
	Document []Data `json:"document"`
	Request  *http.Request
	Inserts  []*Insert
	Values   map[string]interface{} //  параметры переданные в шаблон при генерации страницы (доступны в шаблоне как $.Value)
	App      *app
}

// Вставка - это одна функция, которая может иметь вложения
// Text - строка вставки, по которому мы будем заменять в общем тексте
type Insert struct {
	Text      string      `json:"text"`
	Arguments []string    `json:"arguments"`
	Result    interface{} `json:"result"`
	Functions Function
}

// Исчисляемая фукнция с аргументами и параметрами
// может иметь вложения
type Function struct {
	Name      string   `json:"name"`
	Arguments []string `json:"arguments"`
	Result    string   `json:"result"`
}

////////////////////////////////////////////////////////////
// !!! ПОКА ТОЛЬКО ПОСЛЕДОВАТЕЛЬНАЯ ОБРАБОТКА (без сложений)
////////////////////////////////////////////////////////////

func (p *Formula) Replace() (result string) {

	p.Parse()
	p.Calculate()

	for _, v := range p.Inserts {
		p.Value = strings.Replace(p.Value, v.Text, fmt.Sprint(v.Result), -1)
	}

	return p.Value
}

func (p *Formula) Parse() bool {

	if p.Value == "" {
		return false
	}

	//content := []byte(p.Value)
	//pattern := regexp.MustCompile(`@(\w+)\(([\w]+)(?:,\s*([\w]+))*\)`)
	value := p.Value

	pattern := regexp.MustCompile(`@(\w+)\(\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?(?:,\s*('[^']*'|#[^#]*#|[^,()@]*?)\s*)?\)`)
	allIndexes := pattern.FindAllStringSubmatch(value, -1)

	for _, loc := range allIndexes {

		i := Insert{}
		f := Function{}
		i.Functions = f
		p.Inserts = append(p.Inserts, &i)

		strFunc := string(loc[0])

		strFunc1 := strings.Replace(strFunc, "@", "", -1)
		strFunc1 = strings.Replace(strFunc1, ")", "", -1)
		f1 := strings.Split(strFunc1, "(")

		if len(f1) == 1 { // если не нашли ( значит неверно задана @-фукнций
			return false
		}

		i.Text = strFunc
		i.Functions.Name = f1[0] // название функции

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
			i.Functions.Arguments = argsClear
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
		//		i.Functions.Name = res
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
		//			i.Functions.Arguments = append(i.Functions.Arguments, res)
		//		}
		//	}
		//}
	}

	return true
}

func (p *Formula) Calculate() {

	for k, v := range p.Inserts {
		param := strings.ToUpper(v.Functions.Name)

		switch param {
		case "IMGRESIZE":
			p.Inserts[k].Result = p.App.Sendmail(v.Functions.Arguments)
		case "RAND":
			uuid := funcs.UUID()
			p.Inserts[k].Result = uuid[1:6]
		case "SENDMAIL":
			p.Inserts[k].Result = p.App.Sendmail(v.Functions.Arguments)
		case "PATH":
			p.Inserts[k].Result = p.App.Path(p.Document, v.Functions.Arguments)
		case "REPLACE":
			p.Inserts[k].Result = p.App.DReplace(v.Functions.Arguments)
		case "TIME":
			p.Inserts[k].Result = p.App.Time(v.Functions.Arguments)
		case "DATEMODIFY":
			p.Inserts[k].Result = p.App.DateModify(v.Functions.Arguments)

		case "QUERY":
			r, err := p.App.Query(p.Request, v.Functions.Arguments)
			if err != nil {
				p.Inserts[k].Result = err
			} else {
				p.Inserts[k].Result = r
			}

		case "USER":
			p.Inserts[k].Result = p.App.UserObj(p.Request, v.Functions.Arguments)
		case "ROLE":
			p.Inserts[k].Result = p.App.UserRole(p.Request, v.Functions.Arguments)
		case "PROFILE":
			p.Inserts[k].Result = p.App.UserProfile(p.Request, v.Functions.Arguments)

		case "OBJ":
			p.Inserts[k].Result = p.App.Obj(p.Document, v.Functions.Arguments)
		case "URL":
			p.Inserts[k].Result = p.App.FuncURL(p.Request, v.Functions.Arguments)
		case "COOKIE":
			p.Inserts[k].Result = p.App.Cookie(p.Request, v.Functions.Arguments)

		case "SPLITINDEX":
			p.Inserts[k].Result = p.App.SplitIndex(v.Functions.Arguments)

		case "TPLVALUE":
			p.Inserts[k].Result = p.App.TplValue(p.Values, v.Functions.Arguments)
		case "CONFIGVALUE":
			p.Inserts[k].Result = p.App.ConfigValue(v.Functions.Arguments)
		case "FIELDVALUE":
			p.Inserts[k].Result = p.App.FieldValue(p.Document, v.Functions.Arguments)
		case "FIELDSRC":
			p.Inserts[k].Result = p.App.FieldSrc(p.Document, v.Functions.Arguments)
		case "FIELDSPLIT":
			p.Inserts[k].Result = p.App.FieldSplit(p.Document, v.Functions.Arguments)
		default:
			p.Inserts[k].Result = ""
		}
	}

}

///////////////////////////////////////////////////
// Фукнции @ обработки
///////////////////////////////////////////////////

// Получение значений $.Value шаблона (работает со значением по-умолчанию)
func (c *app) TplValue(v map[string]interface{}, arg []string) (result string) {
	var valueDefault string

	// берем через глобальную переменную, через (c *app) не работает для ф-ций шаблонизатора
	if len(c.ConfigParams()) == 0 {
		return "Error parsing @-function TplValue (State is null)"
	}

	if len(arg) > 0 {
		param := arg[0]
		if len(arg) == 2 {
			valueDefault = arg[1]
		}

		result, found := v[strings.Trim(param, " ")]
		if !found {
			if valueDefault == "" {
				return "Error parsing @-function TplValue (Value from this key is not found.)"
			}
			return fmt.Sprint(valueDefault)
		}
		return fmt.Sprint(result)

	} else {
		return "Error parsing @-function TplValue (Arguments is null)"
	}

	return fmt.Sprint(result)
}

// Получение значений из конфигурации проекта (хранится State в объекте приложение App)
func (c *app) ConfigValue(arg []string) (result string) {
	var valueDefault string

	// берем через глобальную переменную, через (c *app) не работает для ф-ций шаблонизатора
	if len(c.ConfigParams()) == 0 {
		return "Error parsing @-function ConfigValue (State is null)"
	}

	if len(arg) > 0 {
		param := arg[0]
		if len(arg) == 2 {
			valueDefault = arg[1]
		}

		result := c.ConfigGet(strings.Trim(param, " "))
		if result == "" {
			if valueDefault == "" {
				return "Error parsing @-function ConfigValue (Value from this key is not found.)"
			}
			return fmt.Sprint(valueDefault)
		}
		return fmt.Sprint(result)

	} else {
		return "Error parsing @-function ConfigValue (Arguments is null)"
	}

	return fmt.Sprint(result)
}

// Получаем значение из разделенной строки по номер
// параметры:
// str - текст (строка)
// sep - разделитель (строка)
// index - порядковый номер (число) (от 0) возвращаемого элемента
// default - значение по-умолчанию (не обязательно)
func (c *app) SplitIndex(arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {

		str := funcs.Replace(arg[0], "'", "", -1)
		sep := funcs.Replace(arg[1], "'", "", -1)
		index := funcs.Replace(arg[2], "'", "", -1)
		defaultV := funcs.Replace(arg[3], "'", "", -1)

		in, err := strconv.Atoi(index)
		if err != nil {
			result = "Error! Index must be a number."
		}
		if len(arg) == 4 {
			valueDefault = defaultV
		}

		slice_str := strings.Split(str, sep)
		result = slice_str[in]

		//fmt.Println(str)
		//fmt.Println(sep)
		//fmt.Println(in)
		//fmt.Println(slice_str)
	}
	if result == "" {
		result = valueDefault
	}

	//fmt.Println(result)

	return result
}

// Получение текущей даты
func (c *app) Time(arg []string) (result string) {

	if len(arg) > 0 {
		param := strings.ToUpper(arg[0])

		switch param {
		case "NOW", "THIS":
			result = time.Now().Format("2006-01-02 15:04:05")
		default:
			result = time.Now().String()
		}
	}

	return result
}

// Получение идентификатор User-а
func (c *app) TimeFormat(arg []string) (result string) {
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

		result = funcs.timeformat(ss, mask, format)
	}
	if result == "" {
		result = valueDefault
	}

	return result
}

func (c *app) FuncURL(r *http.Request, arg []string) (result string) {
	r.ParseForm()
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

	return result
}

// Вставляем значения системных полей объекта
func (c *app) Path(d []Data, arg []string) (result string) {
	var valueDefault string

	if len(arg) > 0 {
		param := strings.ToUpper(arg[0])

		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

		if len(c.ConfigParams()) != 0 {
			switch param {
			case "API":
				result = c.ConfigGet("url_api")
			case "ORM":
				result = c.ConfigGet("url_orm")
			case "PROXY":
				result = c.ConfigGet("url_proxy")
			case "CLIENT":
				result = c.ConfigGet("client_path")
			case "DOMAIN":
				result = c.ConfigGet("domain")
			default:
				result = c.ConfigGet("client_path")
			}
		}
	}

	if result == "" {
		result = valueDefault
	}

	return result
}

// Заменяем значение
func (c *app) DReplace(arg []string) (result string) {
	var count int
	var str, oldS, newS string
	var err error

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

	return result
}

// Получение идентификатор User-а (для Cockpit-a)
func (c *app) UserObj(r *http.Request, arg []string) (result string) {

	//fmt.Println("User")
	//fmt.Println(arg)

	var valueDefault string

	if len(arg) > 0 {

		if len(arg) == 2 {
			valueDefault = strings.ToUpper(arg[1])
		}

		param := strings.ToUpper(arg[0])
		ctxUser := r.Context().Value("User") // текущий профиль пользователя

		var uu ProfileData

		json.Unmarshal([]byte(funcs.marshal(ctxUser)), &uu)

		if &uu != nil {
			switch param {
			case "UID", "ID":
				result = uu.Uid
			case "PHOTO":
				result = uu.Photo
			case "AGE":
				result = uu.Age
			case "NAME":
				result = uu.First_name + " " + uu.Last_name
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

	return result
}

// Получение UserProfile (для Cockpit-a)
func (c *app) UserProfile(r *http.Request, arg []string) (result string) {
	if len(arg) > 0 {

		param := strings.ToUpper(arg[0])
		ctxUser := r.Context().Value("User") // текущий профиль пользователя

		var uu *ProfileData
		if ctxUser != nil {
			ii, _ := json.Marshal(ctxUser)
			json.Unmarshal(ii, &uu)
		}

		if uu != nil {
			role := uu.CurrentProfile
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

	}

	return result
}

// Получение текущей роли User-а
func (c *app) UserRole(r *http.Request, arg []string) (result string) {
	if len(arg) > 0 {

		param := strings.ToUpper(arg[0])
		param2 := strings.ToUpper(arg[1])
		ctxUser := r.Context().Value("User") // текущий профиль пользователя

		var uu *ProfileData
		if ctxUser != nil {
			ii, _ := json.Marshal(ctxUser)
			json.Unmarshal(ii, &uu)
		}

		if uu != nil {
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

	}
	return result
}

// Cookie Получение cookie
// параметры: 	1-й параметр - name (имя куки)
//
//	2-й параметр - field (поле куки: VALUE, EXPIRES, MAXAGE, PATH, SECURE)
func (c *app) Cookie(r *http.Request, arg []string) (result string) {
	if len(arg) > 0 {

		name := arg[0]
		field := strings.ToUpper(arg[1])

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
	}
	return result
}

// Вставляем значения системных полей объекта
func (c *app) Obj(data []Data, arg []string) (result string) {
	var valueDefault, r string
	var res = []string{}
	if len(arg) == 0 {
		result = "Ошибка в переданных параметрах"
	}

	param := strings.ToUpper(arg[0])
	separator := "," // значение разделителя по-умолчанию

	if len(arg) == 0 {
		return "Ошибка в переданных параметрах."
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
	result = funcs.join(res, separator)

	if result == "" {
		result = valueDefault
	}

	return result
}

// Вставляем значения (Value) элементов из формы
// Если поля нет, то выводит переданное значение (может быть любой символ)
func (c *app) FieldValue(data []Data, arg []string) (result string) {
	var valueDefault, separator string
	var resSlice = []string{}

	separator = "," // значение разделителя по-умолчанию

	if len(arg) == 0 {
		return "Ошибка в переданных параметрах."
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
	result = funcs.join(resSlice, separator)

	if result == "" {
		result = valueDefault
	}

	return result
}

// Вставляем ID-объекта (SRC) элементов из формы
// Если поля нет, то выводит переданное значение (может быть любой символ)
func (c *app) FieldSrc(data []Data, arg []string) (result string) {
	var valueDefault, separator string
	var resSlice = []string{}

	if len(arg) == 0 {
		return "Ошибка в переданных параметрах."
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
	result = funcs.join(resSlice, separator)

	if result == "" {
		result = valueDefault
	}

	return result
}

// Разбиваем значения по элементу (Value(по-умолчанию)/Src) элементов из формы по разделителю и возвращаем
// значение по указанному номеру (начала от 0)
// Синтаксис: FieldValueSplit(поле, элемент, разделитель, номер_элемента)
// для разделителя есть кодовые слова slash - / (нельзя вставить в фукнцию)
func (c *app) FieldSplit(data []Data, arg []string) (result string) {
	var resSlice = []string{}
	var r string

	if len(arg) == 0 {
		return "Ошибка в переданных параметрах."
	}

	if len(arg) < 4 {
		return "Error! Count params must have 4 (field, element, separator, number)"
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
		return fmt.Sprint(err)
	}

	for _, d := range data {
		// 2. получили значение поля
		val, found := d.Attr(field, element)

		if !found {
			return "Error! This field is not found."
		}
		in := strings.Trim(val, " ")
		if sep == "slash" {
			sep = "/"
		}

		// 3. разделили и получили нужный элемент
		split_v := strings.Split(in, sep)
		if len(split_v) < num {
			return "Error! Array size is less than the passed number"
		}

		r = split_v[num]
		resSlice = append(resSlice, r)
	}

	result = funcs.join(resSlice, ",")

	return result
}

///////////////////////////////////////////////////
// Фукнции @ обработки наследованные от математического пакета
///////////////////////////////////////////////////

// Добавление даты к переданной
// date - дата, которую модифицируют (значение должно быть в формате времени)
// modificator - модификатор (например "+24h")
// format - формат переданного времени (по-умолчанию - 2006-01-02T15:04:05Z07:00 (формат: time.RFC3339)
func (c *app) DateModify(arg []string) (result string) {

	if len(arg) < 2 {
		return "Error! Count params must have min 2 (date, modificator; option: format)"
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
		fmt.Println("err: ", err)
		return dateArg
	}

	// преобразуем модификатор во время
	d, err := time.ParseDuration(modificator)
	if err != nil {
		return dateArg
	}

	return fmt.Sprint(date.Add(d))
}

// Отправляем почтового сообщения
func (c *app) Sendmail(arg []string) (result string) {
	if len(arg) < 9 {
		return "Error! Count params must have min 9 (server, port, user, pass, from, to, subject, message, turbo: string)"
	}
	result = funcs.sendmail(arg[0], arg[1], arg[2], arg[3], arg[4], arg[5], arg[6], arg[7], arg[8])

	return result
}

// ImgResize изменяем размер изображение и отдаем новую ссылку на отрендеренный файл
// path - путь к файлу
// widht, height - ожидаемые размеры (если 0, то не меняем)
// arg - задаем форматы обрезки/сжатия и тд
func (c *app) ImgResize(ctx context.Context, path string, widht, height int, arg []string) (result string, err error) {
	pathResized := strings.Split(path, ".")[0]

	// сначала проверяем наличие файла нужного размера в хранилище
	// если нет - ресайзим и сохраняем

	data, _, err := c.vfs.Read(ctx, pathResized)
	if err != nil {
		return "", fmt.Errorf("error ImgResize, err: %s", err)
	}
	if len(data) > 0 {
		return pathResized, nil
	}

	data, mime, err := c.vfs.Read(ctx, path)
	if err != nil {
		return "", fmt.Errorf("error ImgResize, err: %s", err)
	}

	fmt.Printf(string(data), mime, err)

	return result, err
}

// Query Делаем вложенный запрос
// аргументы:
// queryName - первый параметр - имя запрсоа
// mode - тип ответа
//
//	id (по-умолчанию) 	- список UID-ов
//	data 				- []Data
//	response			- полный ответ формате Response
func (c *app) Query(r *http.Request, arg []string) (result interface{}, err error) {
	valueDefault := "id"
	if len(arg) == 0 {
		return "Ошибка в переданных параметрах.", fmt.Errorf("error input param")
	}
	if len(arg) == 2 {
		valueDefault = arg[1]
	}

	urlCurl := c.ConfigGet("url_api") + "/query/" + arg[0]
	objs, err := c.GUIQuery(urlCurl, r)
	if err != nil {
		return "", fmt.Errorf("error GUIQuery, urlCurl: %s, er: %s", urlCurl, err)
	}

	switch valueDefault {
	case "data":
		return objs.Data, nil
	case "response":
		return "objs", nil
	default:
		var resUIDs []string
		var respData []Data

		// если можно привести, значит формат внутреннего запроса и возвращаем список uid
		r, err := json.Marshal(objs.Data)
		if err != nil {
			return "Error Marshal Query", fmt.Errorf("error Marshal, urlCurl: %s, er: %s", urlCurl, err)
		}

		err = json.Unmarshal(r, &respData)
		if err != nil {
			return "Error Unmarshal Query", fmt.Errorf("error Unmarshal, urlCurl: %s, er: %s", urlCurl, err)
		}

		for _, v := range respData {
			resUIDs = append(resUIDs, v.Uid)
		}

		res := funcs.join(resUIDs, ",")
		logger.Debug(context.Background(), "Query", zap.String("res", res))

		return res, err
	}

	return "", err
}

// DogParse Собачья-обработка (поиск в строке @функций и их обработка)
func (c *app) DogParse(p string, r *http.Request, queryData *[]Data, values map[string]interface{}) (result string) {
	s1 := Formula{
		App: c,
	}

	// прогоняем полученную строку такое кол-во раз, сколько вложенных уровней + 1 (для сравнения)
	for {
		s1.Value = p
		s1.Request = r
		s1.Values = values
		s1.Document = *queryData
		res_parse := s1.Replace()

		if p == res_parse {
			result = res_parse
			break
		}
		p = res_parse
	}

	return result
}
