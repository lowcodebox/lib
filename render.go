package lib

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Cacher interface {
	Status() bool
	Write(ctx context.Context, key string, value interface{}, ttl time.Duration, tags []string, description string) (err error)
	WriteFunc(ctx context.Context, key string, source func() (res interface{}, err error), refreshInterval time.Duration) (result interface{}, err error)
	ReadFunc(ctx context.Context, key string, updateTTL bool, source func() (res interface{}, err error), refreshInterval time.Duration) (result interface{}, err error)
	Read(ctx context.Context, key string, updateTTL bool) (value interface{}, status string, expired bool, err error)
	Delete(ctx context.Context, key string) (err error)
	Close()
}

type RenderReq struct {
	Page       string            `json:"page"`
	Params     string            `json:"params"`
	ParamsMap  map[string]string `json:"paramsMap"`
	ParamsForm map[string]string `json:"paramsForm"`
	ParamsArr  []string          `json:"paramsArr"`
	Host       string            `json:"host"`

	RequestRaw *http.Request
	RespWriter http.ResponseWriter
}

type RenderState struct {
	cache Cacher
}

// RenderCtx - объект передается в контекст рендеринга блока
// Тэг lcml говорит о том, что эти поля будут включены в файл блока при конвертации
type RenderCtx struct {
	RenderReq

	Body    string                 `json:"body" lcml:"body"`
	Setting map[string]interface{} `json:"setting" lcml:"setting"`
	Include map[string]interface{} `json:"include" lcml:"include"`
	Inc     bool                   `json:"inc"`

	Name    string    `json:"name" lcml:"name"` // base name of the file
	Size    int64     `json:"size"`             // length in bytes for regular files; system-dependent for others
	Mode    string    `json:"mode"`             // file mode bits
	ModTime time.Time `json:"modTime"`          // modification time
	IsDir   bool      `json:"isDir"`            // abbreviation for Mode().IsDir()
	Sys     any       `json:"sys"`              // underlying data source (can return nil)

	// для совместимости с прошлой версией
	Value         interface{} `json:"value"`
	RequestRaw    *http.Request
	RespWriter    http.ResponseWriter
	Page          map[string]interface{}
	Configuration map[string]struct {
		Type   string
		Source interface{}
	}
	Shema interface{}
}

// ByteToRenderCtx приводим тело из файла/базы к объекту для рендеринга model.RenderCtx
// приводим к формату toml, а далее в объект (сделано дял того, чтобы проще было вводить новые директивы
// просто надо добавить в описание объекта model.RenderCtx новых полей
// директива - в начале файла начинается с # название директивы
func ByteToRenderCtx(body []byte) (res RenderCtx, err error) {
	var line string
	var result string
	var flagParam, flagBody bool
	res = RenderCtx{}

	// читаем по-строкам и вычитываем все первые строки с директивами
	// формируем файл toml

	scanner := bufio.NewScanner(bytes.NewReader(body))

	for scanner.Scan() {
		line = scanner.Text()
		if strings.HasPrefix(line, "//") && !flagParam {
			line = strings.Replace(line, "//", "", 1)
			line = strings.Replace(line, ":", " =", 1)
			splitLine := strings.Split(line, "=")
			value := strings.TrimSpace(splitLine[1])
			env := strings.TrimSpace(splitLine[0])

			if len(splitLine) == 2 {
				if value[0] != 39 && value[0] != 34 {
					line = env + " = " + `"` + value + `"`
				}
			}
			result = result + line + "\n"
			continue
		}
		flagParam = true
		if !flagBody {
			result = result + "Body = " + `"""`
		}
		flagBody = true
		result = result + line + "\n"
	}

	// если нет тела, то строки закончились раньше чем было добавлено поле Body
	// поэтому добавляем это поле
	if !flagBody {
		result = result + "Body = " + `"""`
	}
	result = result + `"""`

	if err := scanner.Err(); err != nil {
		return res, fmt.Errorf("scaning error: %v\n", err)
	}

	// .toml -> rBody
	_, err = toml.Decode(result, &res)
	if err != nil {
		return res, fmt.Errorf("decode page to toml failed. err: %w", err)
	}

	return res, err
}

// RenderCtxToByte возвращаем содержимое блока из объекта
// ВНИМАНИЕ - reflect (желательно не использовать)
func RenderCtxToByte(obj RenderCtx) (body []byte, err error) {
	var result string
	var valueBody string

	vv := reflect.ValueOf(obj)
	for i := 0; i < vv.NumField(); i++ {
		value := vv.Field(i)
		field := vv.Type().Field(i).Name
		typed := vv.Type().Field(i).Type
		tag := vv.Type().Field(i).Tag.Get("lcml")

		// пропускаем поля, которые не указаны тегом lcdp в структуре model.RenderCtx
		// чтобы исключить внутренние поля, которые мы не ходим видеть в файлах блоков
		if tag == "" {
			continue
		}

		if field != "Body" {
			// для приведения map в вид Setting.JSH (а не map[JSH])
			// через рефлексию вычитываем поля и значения
			if strings.Contains(typed.String(), "map[string]interface") {
				b := value.Interface()
				h := b.(map[string]interface{})
				for k, v := range h {
					result = result + "// " + field + "." + k + ": " + fmt.Sprintf("%v", v) + "\n"
				}

				continue
			}

			result = result + "// " + field + ": " + fmt.Sprintf("%v", value) + "\n"
			continue
		}

		valueBody = fmt.Sprintf("%v", value)
	}

	result = result + "\n" + valueBody

	return []byte(result), err
}
