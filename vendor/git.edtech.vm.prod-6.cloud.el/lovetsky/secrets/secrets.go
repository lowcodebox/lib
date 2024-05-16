package secrets

import (
	"context"
	"encoding/base64"
	"regexp"
	"strings"

	contoller "git.edtech.vm.prod-6.cloud.el/fabric/controller-client"
	"git.edtech.vm.prod-6.cloud.el/fabric/lib"
)

var reSecretsCheck = regexp.MustCompile(`(secret_([^" ]+))`)

// ParseSecrets заменяет все строки вида secret_"key" на секреты из контроллера
func ParseSecrets(ctx context.Context, cfgString, controllerURL, projectKey string, cfg interface{}) error {
	client := contoller.New(controllerURL, false, projectKey)

	matches := reSecretsCheck.FindAllStringSubmatch(cfgString, -1)
	// сет замененных ключей
	visited := map[string]struct{}{}
	for _, match := range matches {
		// должно быть 3 вхождения - с кавычками, без кавычек, без префикса
		if len(match) != 3 {
			continue
		}

		placeholder := match[1]
		key := match[2]
		// Проверка, что уже заменили
		if _, found := visited[key]; found {
			continue
		}
		// забрать секрет из контроллера
		value, err := client.GetSecret(ctx, key)
		if err != nil {
			return err
		}
		visited[key] = struct{}{}
		// заменить все вхождения
		cfgString = strings.Replace(cfgString, placeholder, value, -1)
	}
	// Лучше обратно перегнать через ConfigLoad, а не DecodeConfig, чтобы точно были загружены дефолтовые значения
	cfgString = base64.StdEncoding.EncodeToString([]byte(cfgString))
	_, err := lib.ConfigLoad(cfgString, cfg)
	return err
}
