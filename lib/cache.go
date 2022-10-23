package app_lib

import (
	"fmt"

	"github.com/restream/reindexer"
	"net/http"
	"time"
	"strconv"
	"encoding/json"
)


// формируем ключ кеша
func (l *app) SetCahceKey(r *http.Request, p Data) (key, keyParam string)  {
	key2 := ""
	key3 := ""

	// формируем сложный ключ-хеш
	key1, _ := json.Marshal(p.Uid)
	key2 = r.URL.Path // переводим в текст параметры пути запроса (/nedra/user)
	key3 = fmt.Sprintf("%v", r.URL.Query()) // переводим в текст параметры строки запроса (?sdf=df&df=df)

	cache_nokey2, _ := p.Attr("cache_nokey2", "value")
	cache_nokey3, _ := p.Attr("cache_nokey3", "value")

	// учитываем путь и параметры
	if cache_nokey2 == "" && cache_nokey3 == "" {
		key = l.hash(string(key1)) + "_" + l.hash(string(key2)) + "_" + l.hash(string(key3))
	}

	// учитываем только путь
	if cache_nokey2 != "" && cache_nokey3 == "" {
		key = l.hash(string(key1)) + "_" + l.hash(string(key2)) + "_"
	}

	// учитываем только параметры
	if cache_nokey2 == "" && cache_nokey3 != "" {
		key = l.hash(string(key1)) + "_" + "_" + l.hash(string(key3))
	}

	// учитываем путь и параметры
	if cache_nokey2 != "" && cache_nokey3 != "" {
		key = l.hash(string(key1)) + "_" + "_"
	}

	return key, "url:"+key2+"; params:"+key3
}

// key - ключ, который будет указан в кеше
// option - объект блока (запроса и тд) то, где хранится время кеширования
func (l *app) СacheGet(key string, block Data, r *http.Request, page Data, values map[string]interface{}, url string) (string, bool)  {
	var res string
	var rows *reindexer.Iterator

	rows = l.db.Query(l.ConfigGet("namespace")).
		Where("Uid", reindexer.EQ, key).
		ReqTotal().
		Exec()


	// если есть значение, то обязательно отдаем его, но поменяем
	for rows.Next() {
		elem := rows.Object().(*ValueCache)
		res = elem.Value

		flagFresh := Timefresh(elem.Deadtime);

		if flagFresh == "true" {

			// блокируем запись, чтобы другие процессы не стали ее обновлять также
			if elem.Status != "updating" {

				if 	f := l.refreshTime(block); f == 0 {
					return "", false
				}

				// меняем статус
				elem.Status = "updating"
				l.db.Upsert(l.ConfigGet("namespace"), elem)

				// запускаем обновение кеша фоном
				go l.cacheUpdate(key, block, r, page, values, url)
			}
		}

		//fmt.Println("Отдали из кеша")

		return res, true
	}

	//fmt.Println("Нет в кеша")

	return "", false
}


// key - ключ, который будет указан в кеше
// option - объект блока (запроса и тд) то, где хранится время кеширования
// data - то, что кладется в кеш
func (l *app) CacheSet(key string, block Data, page Data, value, url string) bool {
	var valueCache = ValueCache{}
	var deadTime time.Duration

	// если интервал не задан, то не кешируем
	f := l.refreshTime(block)

	//log.Warning("block: ", block)
	if f == 0 {
		return false
	}

	valueCache.Uid = key
	valueCache.Value = value

	deadTime = time.Minute * time.Duration(f)
	dt := time.Now().UTC().Add(deadTime)

	// дополнитлельные ключи для поиска кешей страницы и блока (отдельно)
	var link []string

	link = append(link, page.Uid)
	link = append(link, block.Uid)

	valueCache.Link = link
	valueCache.Url = url
	valueCache.Deadtime = dt.String()
	valueCache.Status = ""

	err := l.db.Upsert(l.ConfigGet("namespace"), valueCache)
	if err != nil {
		//l.Logger.Error(err, "Error! Created cache from is failed! ")
		fmt.Println(err, "Error! Created cache from is failed! ")
		return false
	}

	//fmt.Println("Пишем в кеш")


	return true
}

func (l *app) cacheUpdate(key string, block Data, r *http.Request, page Data, values map[string]interface{}, url string) {

	// получаем контент модуля
	value := l.ModuleBuild(block, r, page, values, false)

	// обновляем кеш
	l.CacheSet(key, block, page, string(value.result), url)

	return
}

func (l *app) refreshTime(options Data) int {

	refresh, _ := options.Attr("cache", "value")
	if refresh == "" {
		return 0
	}

	f, err := strconv.Atoi(refresh)
	if err != nil {
		return 0
	}

	return f
}