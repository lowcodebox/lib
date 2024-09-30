package lib

import (
	"math/rand"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"
)

// UrlUpdater интерфейс над теми, кто реализует функцию обновления прямых ссылок
type UrlUpdater interface {
	// UpdateDirectURLs обновить прямые ссылки по карте, полученной из контроллера
	UpdateDirectURLs(endpoints map[string][]string)
}

type ExecFunc func(url string) (status int, err error)

// UrlManager управляющий ссылками клиента
// Пытается использовать прямые ссылки, полученные в методе UpdateDirectURLs
// В случае ошибок использует longUrl
type UrlManager struct {
	// Хеш всех текущих ссылок, чтобы заново не обновлять, если ссылки не поменялись
	state string
	// "Долгая" ссылка
	longUrl string
	// Прямые ссылки
	directUrls []string
	// имя/версия сервиса, на который хотим ходить
	domain string
	// мьютекс на обновление прямых ссылок
	directMux *sync.RWMutex
}

// NewUrlManager создать нового управляющего ссылками клиента
// longUrl - основная "долгая" ссылка (через нгинксы и контроллер)
// domain - имя/версия сервиса. Для получения ссылок из карты
func NewUrlManager(longUrl, domain string) *UrlManager {
	return &UrlManager{
		longUrl:   longUrl,
		domain:    domain,
		directMux: &sync.RWMutex{},
	}
}

// UpdateDirectURLs обновить прямые ссылки по карте, полученной из контроллера
func (u *UrlManager) UpdateDirectURLs(endpoints map[string][]string) {
	urls := endpoints[u.domain]
	// Проверить, что новые ссылки как-то изменились
	newState := Hash(strings.Join(urls, ""))
	// Нет изменений - не надо ничего менять
	if newState == u.state {
		return
	}
	u.state = newState

	// Перемешать урлы, чтобы все не ходили через первый, если их несколько
	urls = slices.Clone(urls)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	// Записать изменения
	u.directMux.Lock()
	defer u.directMux.Unlock()
	u.directUrls = urls
}

// Exec выполният запрос. Использует прямые ссылки. Если вернул не нулевой статус(кроме 404) - выходит
// Если вернута ошибка или статус 404 и пробуется следующая ссылка
func (u *UrlManager) Exec(f ExecFunc) (status int, err error) {
	u.directMux.RLock()
	defer u.directMux.RUnlock()

	for _, url := range u.directUrls {
		status, err = f(url)
		if status != 0 && status != http.StatusNotFound {
			break
		}
	}

	return f(u.longUrl)
}
