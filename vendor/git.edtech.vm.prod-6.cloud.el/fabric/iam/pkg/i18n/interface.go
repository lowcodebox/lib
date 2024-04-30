package i18n

import "fmt"

type Localization struct {
	Value map[string]string
}

type I18n struct {
	RegError          Localization
	UserNotExist      Localization
	PassFail          Localization
	UserRolesIsEmpty  Localization
	TokenValidateFail Localization
	TokenValidateOK   Localization
}

func (l *Localization) Text() (result string) {
	return l.Value["RU"]
}

func (l *Localization) Error(payload string) (result error) {
	return fmt.Errorf("%s (%s)", l.Value["RU"], payload)
}

func New() I18n {
	var i = I18n{}

	i.RegError = Localization{map[string]string{}}
	i.RegError.Value["RU"] = "Ошибка регистрации. Пользователь с данным email уже зарегистрирован."
	i.RegError.Value["EN"] = "Error. The user was registered earlier."

	i.UserNotExist = Localization{map[string]string{}}
	i.UserNotExist.Value["RU"] = "Ошибка. Пользователь не найден."
	i.UserNotExist.Value["EN"] = "Error. User is not exist."

	i.PassFail = Localization{map[string]string{}}
	i.PassFail.Value["RU"] = "Ошибка. Логин/пароль не вернен."
	i.PassFail.Value["EN"] = "Error. Pass/login is failed."

	i.UserRolesIsEmpty = Localization{map[string]string{}}
	i.UserRolesIsEmpty.Value["RU"] = "Ошибка. У пользователю нет назначенных профилей"
	i.UserRolesIsEmpty.Value["EN"] = "Error. The user has no profiles assigned."

	i.TokenValidateFail = Localization{map[string]string{}}
	i.TokenValidateFail.Value["RU"] = "Ошибка валидации токена"
	i.TokenValidateFail.Value["EN"] = "Error. Validate token is failed."

	i.TokenValidateOK = Localization{map[string]string{}}
	i.TokenValidateOK.Value["RU"] = "Токен валиден"
	i.TokenValidateOK.Value["EN"] = "Token validation successful"

	return i
}
