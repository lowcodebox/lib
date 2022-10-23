package client

import (
	"encoding/json"
	"fmt"

	"git.lowcodeplatform.net/fabric/lib"
	"git.lowcodeplatform.net/fabric/models"
)

// Refresh отправляем старый X-Auth-Access-токен
// получаем X-Auth-Access токен (два токена + текущая авторизационная сессия)
// этот ключ добавляется в куки или сохраняется в сервисе как ключ доступа
// profile - uid-профиля, под которым проводим авторизацию
// expire - признак того, что refresh-токен прийдет просроченный в новом токене
func (o *iam) Refresh(token, profile string, expire bool) (result string, err error) {
	var res models.Response

	_, err = lib.Curl("POST", o.url + "/token/refresh?profile="+profile+"&expire="+fmt.Sprint(expire), token, &res, map[string]string{}, nil)
	if err != nil {
		return result, err
	}

	result = fmt.Sprint(res.Data)

	return result, err
}

func (o *iam) ProfileGet(sessionID string) (result string, err error) {
	var res models.Response

	_, err = lib.Curl("GET", o.url + "/profile/"+sessionID, "", &res, map[string]string{}, nil)
	if err != nil {
		return result, err
	}

	b2, _ := json.Marshal(res.Data)

	return string(b2), err
}

func (o *iam) ProfileList() (result string, err error) {
	var res models.Response

	_, err = lib.Curl("GET", o.url + "/profile/list", "", &res, map[string]string{}, nil)
	if err != nil {
		return result, err
	}

	result = fmt.Sprint(res.Data)

	return result, err
}