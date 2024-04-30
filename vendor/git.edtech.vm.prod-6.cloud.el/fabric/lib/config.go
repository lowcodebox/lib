package lib

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/gommon/color"
)

var warning = color.Red("[Fail]")

// ConfigLoad читаем конфигурации
// получаем только название конфигурации
// 1. поднимаемся до корневой директории
// 2. от нее ищем полный путь до конфига
// 3. читаем по этому пути
func ConfigLoad(config string, pointToCfg interface{}) (err error) {
	var payload string

	if err := envconfig.Process("", pointToCfg); err != nil {
		fmt.Printf("%s Error load default enviroment: %s\n", warning, err)
		err = fmt.Errorf("Error load default enviroment: %s", err)
		return err
	}

	// проверка на длину конфигурационного файла
	// если он больше 100, то скорее всего передали конфигурацию в base64
	if len(config) < 200 {
		// 3.
		if len(config) == 0 {
			return fmt.Errorf("%s", "Error. Configfile is empty.")
		}
		if !strings.Contains(config, ".") {
			config = config + ".cfg"
		}

		// 4. читаем из файла
		payload, err = ReadFile(config)
		if err != nil {
			return fmt.Errorf("Error raed configfile:  (%s), err: %s", config, err)
		}

	} else {
		// пробуем расшифровать из base64
		debase, err := base64.StdEncoding.DecodeString(config)
		if err != nil {
			return fmt.Errorf("Error decode to string from base64 configfile. err: %s", err)
		}
		payload = string(debase)
	}
	err = DecodeConfig(payload, pointToCfg)

	return err
}

// DecodeConfig Читаем конфигурация из строки
func DecodeConfig(configfile string, cfg interface{}) (err error) {
	if _, err = toml.Decode(configfile, cfg); err != nil {
		fmt.Printf("%s Error: %s (configfile: %s)\n", warning, err, configfile)
	}

	return err
}

// SearchConfigDir получаем путь до искомой конфигурации от переданной директории
func searchConfigDir(startDir, configuration string) (configPath string, err error) {
	var nextPath string
	directory, err := os.Open(startDir)
	if err != nil {
		return "", err
	}
	defer directory.Close()

	objects, err := directory.Readdir(-1)
	if err != nil {
		return "", err
	}

	// пробегаем текущую папку и считаем совпадание признаков
	for _, obj := range objects {
		nextPath = startDir + sep + obj.Name()
		if obj.IsDir() {
			dirName := obj.Name()

			// не входим в скрытые папки
			if dirName[:1] != "." {
				configPath, err = searchConfigDir(nextPath, configuration)
				if configPath != "" {
					return configPath, err // поднимает результат наверх
				}
			}
		} else {
			if !strings.Contains(nextPath, "/.") {
				// проверяем только файлы конфигурации (игнорируем .json)
				if strings.Contains(obj.Name(), configuration+".cfg") {
					return nextPath, err
				}
			}
		}
	}

	return configPath, err
}
