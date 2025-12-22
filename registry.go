package lib

import (
	"math/rand"
	"strconv"
	"time"

	"git.lowcodeplatform.net/packages/models"
)

// OptimizePathMesh получает ссылки на сервисы по их экземплярам
// В данный момент пока преобразует в формат хост:порт и перемешивает
// instances - текущие экземпляры сервиса
func OptimizePathMesh(instances []models.Alive) (urls []string) {
	if len(instances) == 0 {
		return urls
	}

	urls = make([]string, len(instances))
	for i := range instances {
		// Сервис можно работать по http или grpc
		port := instances[i].HTTP
		if port == 0 {
			port = instances[i].Grpc
		}
		urls[i] = instances[i].Path + ":" + strconv.Itoa(port)
	}

	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	rnd.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	return urls
}

// OptimizePathMeshDC получает ссылки на сервисы по их экземплярам
// Преобразует в формат хост:порт и кладет в первый массив тех, кто соответсвует заданному dc, во второй - остальных
func OptimizePathMeshDC(instances []models.Alive, dc string) (urlsPriority, urlsFallback []string) {
	if len(instances) == 0 {
		return nil, nil
	}

	urlsPriority = make([]string, 0, len(instances))
	urlsFallback = make([]string, 0, len(instances))
	var url string
	for i := range instances {
		port := instances[i].HTTP
		url = instances[i].Path + ":" + strconv.Itoa(port)

		if instances[i].DC == dc {
			urlsPriority = append(urlsPriority, url)
			continue
		}

		urlsFallback = append(urlsFallback, url)
	}

	return
}
