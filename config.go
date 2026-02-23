package lib

import (
	"embed"
	"encoding/base64"
	"errors"
	"fmt"

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
func ConfigLoad(config string, cfgPointer interface{}) (payload string, err error) {
	var pbyte []byte

	if err := envconfig.Process("", cfgPointer); err != nil {
		fmt.Println(warning, "Unable load default environment:", err)
		err = fmt.Errorf("unable load default environment: %w", err)
		return "", err
	}

	if len(config) == 0 {
		return "", ErrConfig
	}

	// сначала предполагаем что это файл, если ошибка
	// то скорее всего передали конфигурацию в base64

	// возможно передана папка
	// тогда читаем ее и собираем все файлы вместе в один конфиг
	isDir, _ := IsDir(config)

	// директория - читаем данные рекурсивно из всех папок ниже и объединяем
	if isDir {
		var complexFile string
		mapFiles, err := ReadFilesToMap(config, false)
		if err != nil {
			return "", err
		}
		for _, file := range mapFiles {
			complexFile = complexFile + "\n" + string(file)
		}

		err = DecodeConfig(complexFile, cfgPointer)
		return complexFile, err
	}

	// пробуем читаем из файла
	pbyte, err = ReadFile(config)
	if err != nil {
		// пробуем расшифровать из base64
		pbyte, err = base64.StdEncoding.DecodeString(config)
		if err != nil {
			return "", fmt.Errorf("unable unable read configfile/decode to string from base64 configfile (%s): %w", config, err)
		}
	}

	payload = string(pbyte)
	err = DecodeConfig(payload, cfgPointer)

	return payload, err
}

// DecodeConfig Читаем конфигурация из строки
func DecodeConfig(configfile string, cfg interface{}) (err error) {
	if _, err = toml.Decode(configfile, cfg); err != nil {
		fmt.Println(warning, "Error:", err, "(configfile: "+configfile+")")
	}

	return err
}

// IniLocalize - копируем конфиг по-умолчанию в директорию запуска сервера
func IniLocalize(iniFS embed.FS, config string) (err error) {

	// копируем установочные файлы в /ini
	existDir := IsExist("./ini")
	if !existDir {
		err = CreateDir("./ini", 0766)
		if err != nil {
			return fmt.Errorf("create /ini directory failed. err: %w", err)
		}

		// копируем файлы
		dataFile, err := iniFS.ReadFile("ini/config.toml")
		if err != nil {
			return fmt.Errorf("read config file (in iniFS) failed. err: %w", err)
		}

		err = WriteFile(config, dataFile)
		if err != nil {
			return fmt.Errorf("create config file failed. err: %w", err)
		}
	}

	return err
}
