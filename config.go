package lib

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/gommon/color"
)

var (
	ErrConfig = errors.New("config file is empty")
	warning   = color.Red("[Fail]")
)

// ConfigLoad читаем конфигурации
// получаем только название конфигурации
// 1. поднимаемся до корневой директории
// 2. от нее ищем полный путь до конфига
// 3. читаем по этому пути
func ConfigLoad(config, serviceVersion, hashCommit string, cfgPointer interface{}) (payload string, err error) {
	var pbyte []byte
	var configFile string

	if len(config) == 0 {
		return "", ErrConfig
	}

	if err := envconfig.Process("", cfgPointer); err != nil {
		fmt.Println(warning, "Unable load default environment:", err)
		err = fmt.Errorf("unable load default environment: %w", err)
		return "", err
	}

	// сначала предполагаем что это файл, если ошибка
	// то скорее всего передали конфигурацию в base64
	if !strings.Contains(config, ".") {
		configFile = config + ".cfg"
	}

	// 4. читаем из файла
	pbyte, err = ReadFile(configFile)
	if err != nil {
		// пробуем расшифровать из base64
		pbyte, err = base64.StdEncoding.DecodeString(config)
		if err != nil {
			return "", fmt.Errorf("unable unable read configfile/decode to string from base64 configfile: %w", err)
		}
	}

	payload = string(pbyte)

	err = DecodeConfig(payload, cfgPointer)
	_ = DecodeConfig(payload, &pingConf)
	_ = DecodeConfig(payload, &pingConfOld)
	pingConf.Version = serviceVersion
	pingConf.HashCommit = hashCommit
	configName = strings.Split(config, ".")[0]

	return payload, err
}

// DecodeConfig Читаем конфигурация из строки
func DecodeConfig(configfile string, cfg interface{}) (err error) {
	if _, err = toml.Decode(configfile, cfg); err != nil {
		fmt.Println(warning, "Error:", err, "(configfile: "+configfile+")")
	}

	return err
}

// searchConfigDir — получаем путь до искомой конфигурации от переданной директории
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
