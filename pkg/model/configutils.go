package model

import (
	"encoding/json"
	"fmt"
	"git.lowcodeplatform.net/fabric/lib"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const sep = string(filepath.Separator)

//Float32 custom duration for toml configs
type Float struct {
	float64
	Value float64
}

//UnmarshalText method satisfying toml unmarshal interface
func (d *Float) UnmarshalText(text []byte) error {
	var err error
	i, err := strconv.ParseFloat(string(text), 10)
	d.Value = i
	return err
}

//Float32 custom duration for toml configs
type Bool struct {
	bool
	Value bool
}

//UnmarshalText method satisfying toml unmarshal interface
func (d *Bool) UnmarshalText(text []byte) error {
	var err error
	d.Value = false
	if string(text) == "true" {
		d.Value = true
	}
	return err
}

//Duration custom duration for toml configs
type Duration struct {
	time.Duration
	Value time.Duration
}

//UnmarshalText method satisfying toml unmarshal interface
func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	t := string(text)
	// если получили только цифру - добавляем секунды (по-умолчанию)
	if len(t) != 0 {
		lastStr := t[len(t)-1:]
		if lastStr != "h" && lastStr != "m" && lastStr != "s" {
			t = t + "m"
		}
	}
	d.Value, err = time.ParseDuration(t)
	return err
}

//Duration custom duration for toml configs
type Int struct {
	int
	Value int
}

//UnmarshalText method satisfying toml unmarshal interface
func (d *Int) UnmarshalText(text []byte) error {
	var err error
	tt := string(text)
	if tt == "" {
		d.Value = 0
		return nil
	}
	i, err := strconv.Atoi(tt)
	d.Value = i
	return err
}

// формируем ClientPath из Domain
func (c *Config) SetClientPath()  {
	pp := strings.Split(c.Domain, "/")
	name := "buildbox"
	version := "gui"

	if len(pp) == 1 {
		name = pp[0]
	}
	if len(pp) == 2 {
		name = pp[0]
		version = pp[1]
	}
	c.ClientPath = "/" + name + "/" + version

	return
}

// задаем директорию по-умолчанию
func (c *Config) SetRootDir()  {
	rootdir, err := lib.RootDir()
	if err != nil {
		return
	}
	c.RootDir = rootdir
}

// получаем название конфигурации по-умолчанию (стоит галочка=ON)
func (c *Config) SetConfigName()  {
	//fileconfig, err := lib.DefaultConfig()
	//if err != nil {
	//	return
	//}
	//c.ConfigName = fileconfig
}

// получаем значение из конфигурации по ключу
func (c *Config) GetValue(key string) (result string, err error) {
	var rr = map[string]interface{}{}
	var flagOk = false

	// преобразуем значение типа конфигурации в структуру и получем значения в тексте
	b1, _ := json.Marshal(c)
	json.Unmarshal(b1, &rr)

	for i, v := range rr {
		if i == key {
			result = fmt.Sprint(v)
			flagOk = true
		}
	}
	if !flagOk {
		err = fmt.Errorf("%s", "Value from key not found")
	}
	return
}