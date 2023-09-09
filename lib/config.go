package app_lib

import (
	"fmt"
	"runtime/debug"
)

// ConfigGet метод, которые проверяем наличие ключа в стейте приложения и если нет, то пишет в лог
func (s *app) ConfigGet(key string) (value string) {
	s.config.mx.Lock()
	defer s.config.mx.Unlock()

	value, found := s.config.payload[key]
	if !found {
		//err := errors.New("Key '" + key + "' from application state not found")
		//fmt.Println(err)
		//s.Logger.Error(err)
	}
	return value
}

// ConfigSet метод, которые проверяем наличие ключа в стейте приложения и если нет, то пишет в лог
func (s *app) ConfigSet(key, value string) (err error) {
	s.config.mx.Lock()
	defer s.config.mx.Unlock()

	s.config.payload[key] = value
	return err
}

// метод возвращает все ключи
func (s *app) ConfigParams() map[string]string {
	defer func() {
		rec := recover()
		if rec != nil {
			b := string(debug.Stack())
			fmt.Println(b)
		}
	}()

	s.config.mx.Lock()
	defer s.config.mx.Unlock()

	return s.config.payload
}
