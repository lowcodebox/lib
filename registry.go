package lib

import (
	"math/rand"
	"strconv"
	"time"

	"git.edtech.vm.prod-6.cloud.el/fabric/models"
)

// OptimizePathMesh получает ссылки на сервисы по их экземплярам
// В данный момент пока преобразует в формат хост:порт и перемешивает
// instances - текущие экземпляры сервиса
func OptimizePathMesh(instances []models.Pong) (urls []string) {
	if len(instances) == 0 {
		return urls
	}

	urls = make([]string, len(instances))
	for i := range instances {
		// Сервис можно работать по http или grpc
		port := instances[i].PortHTTP
		if port == 0 {
			port = instances[i].PortGrpc
		}
		urls[i] = instances[i].Host + ":" + strconv.Itoa(port)
	}

	rnd := rand.New(rand.NewSource(time.Now().Unix()))
	rnd.Shuffle(len(urls), func(i, j int) {
		urls[i], urls[j] = urls[j], urls[i]
	})

	return urls
}
