// TokenACL - токен для определения прав доступа к объекту
// формат map[uid_субьекта_доступа]битовая маска
// битовая маска состоит из 2 байт (16 бит)
// субьекты доступа - сущности (объекты), от лица которых получается доступ к ресурсу (объекту)
// существуют 4 вида субьектов - User/Role/Group/Others (URGO)
// правила субьектов применяются исходя из приоритета
// самый приоритетный User, далее по-нисходящей
// право субьекта кодируется двумя битами
// первый бит - статус Deny, второй - статус Access
// бит запрета (Deny) приоритетнее
// если для доступа к объекту мы получаем запрос от нескольких субьектов в рамках одного типа,
// то при наличии Access=1 хоть у одного из них - Доступ разрешен
// наличие Deny = 1, хоть у одного из них - Доступ запрещен
// более приоритетные типы субьектов могут изменить доступ, если они заданы
// 00 - состояние, при котором для данного типа субьекта нет запроса на права (учитываются другие состояния)
// 01 - состояние = Доступ разрешен
// 10 - состояние  = Доступ запрещен
// 11 - конфликт (возможно если у одного субьекта доступ Разрешен, а для другого Запрещен) - исключительная ситуация = Доступ запрещен
// После расчета назначенных прав в рамках всех типов субьектов, права применяются согласно приоритета и мы получаем суммарный статус
// Если на объект нет запроса на доступ или суммарное состояние = 00 - это означает что никто явно СУММАРНО не запросил доступ = Доступ запрещен
// Данные операции производятся для каждого из прав Read/Write/Execute/Admin

package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"git.lowcodeplatform.net/packages/models"
)

const (
	ACLPermissionAllow    = "allow"
	ACLPermissionDeny     = "deny"
	ACLPermissionNull     = "null"
	ACLPermissionConflict = "conflict"

	ACLPriorityOthers = "others"
	ACLPriorityRole   = "role"
	ACLPriorityGroup  = "group"
	ACLPriorityUser   = "user"
)

var ErrNoACLKey = errors.New("empty acl key")
var ErrInvalidACLKey = errors.New("acl token is invalid")
var ErrExpiredACLKey = errors.New("acl token is expired")

// CreateACLValue создаем список прав для заданного субъекта
func CreateACLValue(priority string, read, write, execute, admin string) (code uint16, err error) {
	bytePermission := EncodeToByte(read, write, execute, admin)
	bytePriority := EncodePriority(priority)

	code = uint16(bytePriority)<<8 | uint16(bytePermission)

	return code, nil
}

// GenTokenACL создаем токен
// ACL - список субъектов с правами доступа (см. описание models.XACLKey)
func GenTokenACL(acl map[string]uint16, projectKey []byte, uid string, tokenInterval time.Duration) (token string, err error) {
	if tokenInterval == 0 {
		tokenInterval = 100000 * time.Hour // 12 лет
	}
	t := models.TokenACL{
		ACL:     acl,
		Expired: time.Now().Add(tokenInterval).Unix(),
		Uid:     uid,
	}

	return encodeACLKey(t, projectKey)
}

// ParseTokenACL - возвращаем права для заданного объекта для данного токена
// ACL - список субьектов с правами доступа (см. описание models.TokenACL)
// token - токен, по-которому проверяем доступ для запроса со списком subjects
// subjects - список субьектов доступа к объекту (чей токен) - списком через запятую (,)
// r, w, x, a - read/write/execute/admin
func ParseTokenACL(token, uid, subjects string, projectKey []byte) (r, w, x, a bool, err error) {
	var res = map[string][4]string{} // ключ - приоритет, значения - права

	valid, acl, err := verifyTokenACL(token, uid, projectKey)
	if !valid || err != nil {
		return false, false, false, false, err
	}

	// пробегаем переданных субъекты, ищем каждый из них в списке доступа и рассчитываем права
	for _, v := range strings.Split(subjects, ",") {
		// если в acl есть указанный пользователь - сохраняем его права
		if rights, ok := acl[v]; ok {
			var currentPer = [4]string{}
			var found bool

			// делим значение на два байта
			priorityByte, permissionByte := byte(rights>>8), byte(rights&0xFF)

			// получаем права на объект (текстовом виде)
			permission := DecodeFromByte(permissionByte)
			priority := DecodePriority(priorityByte)

			if currentPer, found = res[priority]; !found {
				res[priority] = [4]string{}
			}
			for i := 0; i < len(currentPer); i++ {
				// пропускаем, если нет значения, которое могло бы поменять статус
				if permission[i] == "" {
					continue
				}
				// пропускаем если для данного приоритета уже задано Deny
				if currentPer[i] == ACLPermissionDeny {
					continue
				}
				// Не обнуляем, а заменяем следующим значением
				if permission[i] != ACLPermissionNull {
					currentPer[i] = permission[i]
				}
			}
			res[priority] = currentPer
		}
	}

	// складываем права согласно приоритетам
	// тут я могу заменять запрещающие права на более высоком приоритете
	finalPermission := [4]string{}
	for _, pr := range []string{ACLPriorityOthers, ACLPriorityRole, ACLPriorityGroup, ACLPriorityUser} {
		for i := 0; i < 4; i++ {
			if res[pr][i] != ACLPermissionNull && res[pr][i] != "" {
				finalPermission[i] = res[pr][i]
			}
		}
	}

	// если права не заданы - то разрешаем false
	return finalPermission[0] == ACLPermissionAllow, finalPermission[1] == ACLPermissionAllow, finalPermission[2] == ACLPermissionAllow, finalPermission[3] == ACLPermissionAllow, err
}

// verifyTokenACL берем X-ACL-Key. если он есть, то он должен быть расшифровать и валидируем содержимое
// acl - возвращает ACL-лист
func verifyTokenACL(token, uid string, projectKey []byte) (valid bool, acl map[string]uint16, err error) {
	xsKey, err := decodeACLKey(projectKey, token)
	if err != nil {
		return false, nil, ErrInvalidACLKey
	}

	if time.Now().Unix() > xsKey.Expired {
		return false, nil, ErrExpiredACLKey
	}

	if xsKey.Uid != uid {
		return false, nil, ErrInvalidACLKey
	}

	return true, xsKey.ACL, nil
}

// decodeACLKey расшифровывает модель XACLKey из токена
// если токен протух - возвращаем ошибку
func decodeACLKey(projectKey []byte, xACLKey string) (xsKey models.TokenACL, err error) {
	if xACLKey == "" {
		return xsKey, ErrNoACLKey
	}

	v, err := Decrypt(projectKey, xACLKey)
	err = json.Unmarshal([]byte(v), &xsKey)
	return
}

// encodeACLKey шифрует модель XACLKey в токен
func encodeACLKey(xsKey models.TokenACL, projectKey []byte) (token string, err error) {
	strJson, err := json.Marshal(xsKey)
	if err != nil {
		return "", fmt.Errorf("error Marshal XACLKey, err: %s", err)
	}

	token, err = Encrypt(projectKey, string(strJson))
	if err != nil {
		return "", fmt.Errorf("error Encrypt XACLKey, err: %s", err)
	}
	return token, nil
}

// EncodeToByte Кодирование 4 значений в байт
func EncodeToByte(v1, v2, v3, v4 string) byte {
	encode := func(v string) byte {
		switch v {
		case ACLPermissionAllow:
			return 0b01
		case ACLPermissionDeny:
			return 0b10
		case ACLPermissionNull:
			return 0b00
		case ACLPermissionConflict:
			return 0b11
		default:
			return 0b00
		}
	}

	return encode(v1)<<6 | encode(v2)<<4 | encode(v3)<<2 | encode(v4)
}

// DecodeFromByte Декодирование байта в 4 значения
func DecodeFromByte(b byte) [4]string {
	decode := func(bits byte) string {
		switch bits & 0b11 {
		case 0b00:
			return ACLPermissionNull
		case 0b01:
			return ACLPermissionAllow
		case 0b10:
			return ACLPermissionDeny
		case 0b11:
			return ACLPermissionConflict
		default:
			return ACLPermissionNull
		}
	}

	return [4]string{
		decode(b >> 6),
		decode(b >> 4),
		decode(b >> 2),
		decode(b),
	}
}

// EncodePriority - кодируем/декодируем типы субьектов доступа
func EncodePriority(t string) byte {
	switch t {
	case ACLPriorityOthers:
		return 0b00
	case ACLPriorityRole:
		return 0b01
	case ACLPriorityGroup:
		return 0b10
	case ACLPriorityUser:
		return 0b11
	default:
		return 0b00
	}
}

func DecodePriority(code byte) string {
	switch code & 0b11 {
	case 0b00:
		return ACLPriorityOthers
	case 0b01:
		return ACLPriorityRole
	case 0b10:
		return ACLPriorityGroup
	case 0b11:
		return ACLPriorityUser
	default:
		return ACLPriorityOthers
	}
}
