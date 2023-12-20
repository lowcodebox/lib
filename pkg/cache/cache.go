package cache

//import "C"
import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"git.lowcodeplatform.net/fabric/app/pkg/function"
	"git.lowcodeplatform.net/fabric/app/pkg/model"
	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/packages/logger"
	"github.com/labstack/gommon/color"
	"go.uber.org/zap"

	"github.com/restream/reindexer"
)

type cache struct {
	DB       *reindexer.Reindexer
	cfg      model.Config
	function function.Function
	active   bool `json:"active"`
}

type Cache interface {
	Active() bool
	GenKey(uid, path, query string, addСonditionPath, addСonditionURL bool) (key, cacheParams string)
	SetStatus(key, status string) (err error)
	Read(key string) (result, status string, fresh bool, err error)
	Write(key, cacheParams string, cacheInterval int, blockUid, pageUid string, value string) (err error)
	Clear(links string) (count int, err error)
}

// проверяем статус соединения с базой
func (c *cache) Active() bool {
	if c.active {
		return true
	}
	return false
}

// формируем ключ кеша
// addСonditionPath, addСonditionURL - признаки добавления хеша пути и/или запроса в ключе (указываются в кеше блока)
func (c *cache) GenKey(uid, path, query string, addСonditionPath, addСonditionURL bool) (key, cacheParams string) {
	key2 := ""
	key3 := ""

	// формируем сложный ключ-хеш
	key1, _ := json.Marshal(uid)
	key2 = path                     // переводим в текст параметры пути запроса (/nedra/user)
	key3 = fmt.Sprintf("%v", query) // переводим в текст параметры строки запроса (?sdf=df&df=df)
	keyParams := ""

	// учитываем путь и параметры
	if addСonditionPath && addСonditionURL {
		key = lib.Hash(string(key1)) + "_" + lib.Hash(key2) + "_" + lib.Hash(key3)
		keyParams = "url:" + key2 + "; params:" + key3
	}

	// учитываем только путь
	if !addСonditionPath && addСonditionURL {
		key = lib.Hash(string(key1)) + "_" + lib.Hash(key2) + "_"
		keyParams = "url:" + key2 + "; params:"
	}

	// учитываем только параметры
	if addСonditionPath && !addСonditionURL {
		key = lib.Hash(string(key1)) + "_" + "_" + lib.Hash(key3)
		keyParams = "url: ; params:" + key3

	}

	// учитываем путь и параметры
	if !addСonditionPath && !addСonditionURL {
		key = lib.Hash(string(key1)) + "_" + "_"
		keyParams = "url: ; params:"
	}

	return key, keyParams
}

// SetStatus меняем статус по ключу
// для решения проблему дублированного обновления кеша первый, кто инициирует обновление кеша меняет статус на updated
// и меняет время в поле Deadtime на время изменения статуса + 2 минуты = максимальное время ожидания обновления кеша
// это сделано для того, чтобы не залипал кеш, у которых воркер который решил его обновить Отвалился, или был передернут сервис
// таким образом, запрос, который получает старый кеш у которого статут updated проверяем время старта обновления и если оно просрочено
// то сам инициирует обновление кеша (меняя время на свое)
func (c *cache) SetStatus(key, status string) (err error) {
	var rows *reindexer.Iterator
	var deadTime = time.Now().UTC().Add(c.cfg.TimeoutCacheGenerate.Value) // время, когда статус updated перестанет быть валидным

	rows = c.DB.Query(c.cfg.Namespace).
		Where("Uid", reindexer.EQ, key).
		ReqTotal().
		Exec()

	// если есть значение, то обязательно отдаем его, но поменяем
	for rows.Next() {
		elem := rows.Object().(*model.ValueCache)

		// меняем статус
		elem.Status = status
		if status == "updated" {
			elem.Deadtime = deadTime.String()
		}
		err = c.DB.Upsert(c.cfg.Namespace, elem)
	}

	rows.Close()
	return
}

// key - ключ, который будет указан в кеше
// получаем:
// result, status - результат и статус (текст)
// fresh - признак того, что данные актуальны (свежие)
func (c *cache) Read(key string) (result, status string, flagExpired bool, err error) {
	var rows *reindexer.Iterator

	rows = c.DB.Query(c.cfg.Namespace).
		Where("Uid", reindexer.EQ, key).
		ReqTotal().
		Exec()

	// если есть значение, то обязательно отдаем его, но поменяем
	for rows.Next() {
		elem := rows.Object().(*model.ValueCache)
		result = elem.Value

		// функция Timefresh показывает пора ли обновить время (не признак свежести, а наоборот)
		// оставил для совместимости со сторыми версиями
		flagExpired = c.function.TplFunc().TimeExpired(elem.Deadtime)
	}

	// если ничего не нашли, то выдаем ошибку, чтобы сгенерировать кеш
	if rows.TotalCount() == 0 {
		err = fmt.Errorf("%s", "Error. Result is null")
	}

	rows.Close()
	return
}

// key - ключ, который будет указан в кеше
// cacheInterval - время хранени кеша
// blockUid, pageUid - ид-ы блока и страницы (для формирования возможности выборочного сброса кеша)
// data - то, что кладется в кеш
func (c *cache) Write(key, cacheParams string, cacheInterval int, blockUid, pageUid string, value string) (err error) {
	ctx := context.Background()
	var valueCache = model.ValueCache{}
	var deadTime time.Duration

	// интервал не указан - значит не кешируем (не пишем в кеш)
	if cacheInterval == 0 {
		return fmt.Errorf("%s", "Cache interval is empty")
	}

	valueCache.Uid = key
	valueCache.Value = value
	deadTime = time.Minute * time.Duration(cacheInterval)
	dt := time.Now().UTC().Add(deadTime)

	// дополнитлельные ключи для поиска кешей страницы и блока (отдельно)
	var link []string

	link = append(link, pageUid)
	link = append(link, blockUid)

	valueCache.Link = link
	valueCache.Url = cacheParams
	valueCache.Deadtime = dt.String()
	valueCache.Status = ""

	err = c.DB.Upsert(c.cfg.Namespace, valueCache)
	if err != nil {
		logger.Error(ctx, "Error! Created cache from is failed!", zap.Error(err))
		return fmt.Errorf("%s", "Error! Created cache from is failed!")
	}

	return
}

// очищаем кеш приложения по заданному критерия (наличия значения в массиве линков)
func (c *cache) Clear(links string) (count int, err error) {
	if links == "all" {
		// паременты не переданы - удаляем все объекты в заданном неймспейсе
		c.DB.Query(c.cfg.Namespace).Not().WhereString("Uid", reindexer.EQ, "").Delete()
	} else {
		// паременты не переданы - удаляем согласно шаблону
		for _, v := range strings.Split(links, ",") {
			int, err := c.DB.Query(c.cfg.Namespace).Where("Link", reindexer.SET, v).Delete()
			if err != nil {
				err = fmt.Errorf("Error cleaning cache. Now deleted %s objects. Error: ", count, err)
				return count, err
			}
			count = count + int
		}
	}

	return
}

func New(cfg model.Config, function function.Function) Cache {
	ctx := context.Background()
	done := color.Green("[OK]")
	fail := color.Red("[Fail]")
	var cach = cache{
		cfg:      cfg,
		function: function,
	}

	// включено кеширование
	if cfg.CachePointsrc != "" {
		cach.DB = reindexer.NewReindex(cfg.CachePointsrc)
		err := cach.DB.OpenNamespace(cfg.Namespace, reindexer.DefaultNamespaceOptions(), model.ValueCache{})
		if err != nil {
			fmt.Printf("%s Error connecting to database. Plaese check this parameter in the configuration. %s\n", fail, cfg.CachePointsrc)
			fmt.Printf("%s\n", err)
			logger.Error(ctx, fmt.Sprintf("Error connecting to database. Plaese check this parameter in the configuration: %s", cfg.CachePointsrc), zap.Error(err))
			return &cach
		} else {
			fmt.Printf("%s Cache-service is running\n", done)
			logger.Info(ctx, "Cache-service is running")
			cach.active = true
		}
	}

	return &cach
}
