package models

import (
	"encoding/json"
	"errors"
)

var StatusCode = RStatus{
	"OK":                       {"", "Запрос выполнен", 200, "", nil},
	"OKLicenseActivation":      {"", "Лицензия была активирована", 200, "", nil},
	"Unauthorized":             {"", "Ошибка авторизации", 401, "", nil},
	"NotCache":                 {"", "Доступно только в Турбо-режиме", 200, "", nil},
	"NotStatus":                {"", "Ответ сервера не содержит статус выполнения запроса", 501, "", nil},
	"NotExtended":              {"", "На сервере отсутствует расширение, которое желает использовать клиент", 501, "", nil},
	"ErrorFormatJson":          {"", "Ошибка формата JSON-запроса", 500, "ErrorFormatJson", nil},
	"ErrorTransactionFalse":    {"", "Ошибка выполнения тразакции SQL", 500, "ErrorTransactionFalse", nil},
	"ErrorBeginDB":             {"", "Ошибка подключения к БД", 500, "ErrorBeginDB", nil},
	"ErrorPrepareSQL":          {"", "Ошибка подготовки запроса SQL", 500, "ErrorPrepareSQL", nil},
	"ErrorNullParameter":       {"", "Ошибка! Не передан параметр", 503, "ErrorNullParameter", nil},
	"ErrorQuery":               {"", "Ошибка запроса на выборку данных", 500, "ErrorQuery", nil},
	"ErrorScanRows":            {"", "Ошибка переноса данных из запроса в объект", 500, "ErrorScanRows", nil},
	"ErrorNullFields":          {"", "Не все поля заполнены", 500, "ErrorScanRows", nil},
	"ErrorAccessType":          {"", "Ошибка доступа к элементу типа", 500, "ErrorAccessType", nil},
	"ErrorGetData":             {"", "Ошибка доступа данным объекта", 500, "ErrorGetData", nil},
	"ErrorRevElement":          {"", "Значение было изменено ранее.", 409, "ErrorRevElement", nil},
	"ErrorForbiddenElement":    {"", "Значение занято другим пользователем.", 403, "ErrorForbiddenElement", nil},
	"ErrorUnprocessableEntity": {"", "Необрабатываемый экземпляр", 422, "ErrorUnprocessableEntity", nil},
	"ErrorNotFound":            {"", "Значение не найдено", 404, "ErrorNotFound", nil},
	"ErrorReadDir":             {"", "Ошибка чтения директории", 403, "ErrorReadDir", nil},
	"ErrorReadConfigDir":       {"", "Ошибка чтения директории конфигураций", 403, "ErrorReadConfigDir", nil},
	"errorOpenConfigDir":       {"", "Ошибка открытия директории конфигураций", 403, "errorOpenConfigDir", nil},
	"ErrorReadConfigFile":      {"", "Ошибка чтения файла конфигураций", 403, "ErrorReadConfigFile", nil},
	"ErrorReadLogFile":         {"", "Ошибка чтения файла логирования", 403, "ErrorReadLogFile", nil},
	"ErrorScanLogFile":         {"", "Ошибка построчного чтения файла логирования", 403, "ErrorScanLogFile", nil},
	"ErrorPortBusy":            {"", "Указанный порт занят", 403, "ErrorPortBusy", nil},
	"ErrorGone":                {"", "Объект был удален ранее", 410, "ErrorGone", nil},
	"ErrorShema":               {"", "Ошибка формата заданной схемы формирования запроса", 410, "ErrorShema", nil},
	"ErrorInitBase":            {"", "Ошибка инициализации новой базы данных", 410, "ErrorInitBase", nil},
	"ErrorCreateCacheRecord":   {"", "Ошибка создания объекта в кеше", 410, "ErrorCreateCacheRecord", nil},
	"ErrorUpdateParams":        {"", "Не переданы параметры для обновления серверов (сервер источник, сервер получатель)", 410, "ErrorUpdateParams", nil},
	"ErrorIntervalProxy":       {"", "Ошибка переданного интервала (формат: 1000:2000)", 410, "ErrorIntervalProxy", nil},
	"ErrorReservPortProxy":     {"", "Ошибка выделения порта proxy-сервером", 410, "ErrorReservPortProxy", nil},
}

type RStatus map[string]RestStatus
type RestStatus struct {
	Source      string `json:"source,omitempty"`
	Description string `json:"description,omitempty"`
	Status      int    `json:"status,omitempty"`
	Code        string `json:"code,omitempty"`
	Error       error  `json:"error,omitempty"`
}

func (r *RestStatus) MarshalJSON() ([]byte, error) {
	type RestStatusJson struct {
		Source      string `json:"source,omitempty"`
		Description string `json:"description"`
		Status      int    `json:"status"`
		Code        string `json:"code"`
		Error       string `json:"error"`
	}

	errStr := ""
	if r.Error != nil {
		errStr = r.Error.Error()
	}

	return json.Marshal(RestStatusJson{
		Source:      r.Source,
		Description: r.Description,
		Status:      r.Status,
		Code:        r.Code,
		Error:       errStr,
	})
}

func (r *RestStatus) UnmarshalJSON(b []byte) error {
	type RestStatusJson struct {
		Source      string `json:"source"`
		Description string `json:"description"`
		Status      int    `json:"status"`
		Code        string `json:"code"`
		Error       string `json:"error"`
	}
	t := RestStatusJson{}

	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}

	r.Source = t.Source
	r.Description = t.Description
	r.Code = t.Code
	r.Status = t.Status
	if t.Error == "" {
		r.Error = nil
	} else {
		r.Error = errors.New(t.Error)
	}

	return nil
}
