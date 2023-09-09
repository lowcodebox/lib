package types

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// JSON предоставляет тип данных для логирование JSON-строк через zap-логгер с учетом
// маскирования полей с приватными данными.
func JSON(key string, val string) zap.Field {
	hashedVal, err := MaskSensitiveJSONFields(val, nil, nil, nil)
	if err != nil {
		hashedVal = fmt.Sprintf(`{"error": "%s", "val": "%s"}`,
			strings.ReplaceAll(val, `"`, `\"`),
			strings.ReplaceAll(err.Error(), `"`, `\"`),
		)
	}

	return zap.Field{Key: key, Type: zapcore.StringType, String: hashedVal, Integer: 0, Interface: nil}
}

const (
	unknown = iota
	object
	objectEnd
	fieldName
	fieldValue
	array
	arrayValue
	arrayEnd
)

// MaskSensitiveJSONFields маскирует приватные данные в JSON-строках.
func MaskSensitiveJSONFields(jsonString string, excludeKeys, hideKeys, excludeValues []string) (string, error) {
	decoder := json.NewDecoder(strings.NewReader(jsonString))
	decoder.UseNumber()

	strBuilder := new(strings.Builder)

	var (
		token         json.Token
		stateStack    intStack
		lastFieldName string
		err           error
	)

	state := unknown
	beforeNextValue := func() error {
		switch state {
		case fieldValue:
			if err := strBuilder.WriteByte(','); err != nil {
				return errors.Wrap(err, "logger: json mask encoder")
			}

			state = fieldName
		case fieldName:
			if err := strBuilder.WriteByte(':'); err != nil {
				return errors.Wrap(err, "logger: json mask encoder")
			}

			state = fieldValue
		case arrayValue:
			if err := strBuilder.WriteByte(','); err != nil {
				return errors.Wrap(err, "logger: json mask encoder")
			}
		case array:
			state = arrayValue
		case object:
			state = fieldName
		}

		return nil
	}

	pushState := func() error {
		switch state {
		case object, fieldValue:
			stateStack = stateStack.push(fieldValue)
		case array, arrayValue:
			stateStack = stateStack.push(arrayValue)
		case unknown:
			stateStack = stateStack.push(unknown)
		default:
			return fmt.Errorf("logger: json mask encoder: invalid state (json is invalid?): %d", state)
		}

		return nil
	}

	setState := func(newState int) error {
		switch newState {
		case object:
			if err := beforeNextValue(); err != nil {
				return err
			}

			if err := strBuilder.WriteByte('{'); err != nil {
				return errors.Wrap(err, "logger: json mask encoder")
			}

			if err := pushState(); err != nil {
				return err
			}

			state = newState
		case objectEnd:
			if err := strBuilder.WriteByte('}'); err != nil {
				return errors.Wrap(err, "logger: json mask encoder")
			}

			stateStack, state = stateStack.pop()
		case array:
			if err := beforeNextValue(); err != nil {
				return err
			}

			if err := strBuilder.WriteByte('['); err != nil {
				return errors.Wrap(err, "logger: json mask encoder")
			}

			if err := pushState(); err != nil {
				return err
			}

			state = newState
		case arrayEnd:
			if err := strBuilder.WriteByte(']'); err != nil {
				return errors.Wrap(err, "logger: json mask encoder")
			}

			stateStack, state = stateStack.pop()
		}

		return nil
	}

	excludeKeys = append(excludeKeys, defaultExcludeKeys...)
	hideKeys = append(hideKeys, defaultHideKeys...)

	for {
		token, err = decoder.Token()
		if err != nil {
			break
		}

		if v, ok := token.(json.Delim); ok {
			switch v.String() {
			case "{":
				if err := setState(object); err != nil {
					return "", err
				}
			case "}":
				if err := setState(objectEnd); err != nil {
					return "", err
				}
			case "[":
				if err := setState(array); err != nil {
					return "", err
				}
			case "]":
				if err := setState(arrayEnd); err != nil {
					return "", err
				}
			}
		} else {
			if err := beforeNextValue(); err != nil {
				return "", err
			}
			var jsonBytes []byte
			switch dataType := token.(type) {
			case string:
				currentFieldName := ""
				if state == fieldName {
					lastFieldName = dataType
				} else if state == fieldValue {
					currentFieldName = lastFieldName
				}
				jsonBytes, err = json.Marshal(hashSensitiveValue(currentFieldName, dataType, excludeKeys, hideKeys, excludeValues))
			default:
				jsonBytes, err = json.Marshal(dataType)
			}
			if err != nil {
				return "", errors.Wrap(err, "logger: json mask encoder")
			}
			if _, err = strBuilder.Write(jsonBytes); err != nil {
				return "", errors.Wrap(err, "logger: json mask encoder")
			}
		}
	}

	if !errors.Is(err, io.EOF) {
		return jsonString, nil
	}

	return strBuilder.String(), nil
}

func hashSensitiveValue(fieldName, src string, excludeKeys, hideKeys, excludeValues []string) string {
	fieldName = strings.ToLower(fieldName)

	for _, hideKey := range hideKeys {
		if strings.Contains(fieldName, hideKey) {
			return Hide(src)
		}
	}

	for _, excludeKey := range excludeKeys {
		if strings.Contains(fieldName, excludeKey) {
			return Mask(src)
		}
	}

	for _, excludeValue := range excludeValues {
		if strings.HasPrefix(src, excludeValue) {
			return Mask(src)
		}
	}

	return src
}

type intStack []int

func (s intStack) push(v int) intStack {
	return append(s, v)
}

func (s intStack) pop() (intStack, int) {
	if len(s) == 0 {
		return s, 0
	}

	l := len(s)

	return s[:l-1], s[l-1]
}
