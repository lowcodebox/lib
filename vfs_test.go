package lib

import (
	"context"
	"log"
	"testing"
	"time"
)

// пишем вручную в локальную диреторию (смотрим что поменялся контент)
func TestParallelWR(t1 *testing.T) {
	ctx := context.Background()
	var err error

	paramsLocal := make(map[string]string)
	paramsLocal["Endpoint"] = "./_test"
	paramsLocal["Bucket"] = "local"

	clientLocal := NewVfs(
		"Local",
		paramsLocal["Endpoint"],
		"",
		"",
		"",
		paramsLocal["Bucket"],
		"", "",
	)
	if err != nil {
		log.Fatal("failed create git client:", err)
	}

	// читаем из клиента с интервалом 1 сек
	go func() {
		for {
			time.Sleep(1 * time.Second)

			// Получаем
			res, _, err := clientLocal.Read(ctx, "testfile", false)
			if err != nil {
				log.Printf("read error: %v\n", err)
			}

			log.Printf("read from rep. body: %s\n", res)
		}
	}()

	time.Sleep(600 * time.Second)
}
