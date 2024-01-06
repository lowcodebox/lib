package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Float custom duration for toml configs
type Float struct {
	float64
	Value float64
}

// UnmarshalText method satisfying toml unmarshal interface
func (d *Float) UnmarshalText(text []byte) error {
	var err error
	i, err := strconv.ParseFloat(string(text), 10)
	d.Value = i
	return err
}

// Bool custom duration for toml configs
type Bool struct {
	bool
	Value bool
}

// UnmarshalText method satisfying toml unmarshal interface
func (d *Bool) UnmarshalText(text []byte) error {
	var err error
	d.Value = false
	if string(text) == "true" {
		d.Value = true
	}
	return err
}

// Duration custom duration for toml configs
type Duration struct {
	time.Duration
	Value time.Duration
}

// UnmarshalText method satisfying toml unmarshal interface
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

// Int custom duration for toml configs
type Int struct {
	int
	Value int
}

// UnmarshalText method satisfying toml unmarshal interface
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

// GetValue получаем значение из конфигурации по ключу
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
