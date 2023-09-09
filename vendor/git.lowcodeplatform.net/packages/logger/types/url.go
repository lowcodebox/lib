package types

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func URL(key string, val string) zap.Field {
	hashedVal, err := MaskSensitiveURLFields(val, nil, nil)
	if err != nil {
		hashedVal = fmt.Sprintf(`{"error": "%s", "val": "%s"}`,
			strings.ReplaceAll(val, `"`, `\"`),
			strings.ReplaceAll(err.Error(), `"`, `\"`),
		)
	}

	return zap.Field{Key: key, Type: zapcore.StringType, String: hashedVal, Integer: 0, Interface: nil}
}

func MaskSensitiveURLFields(urlString string, excludeKeys, hideKeys []string) (string, error) {
	var (
		query  url.Values
		parsed *url.URL
		err    error
	)

	isParamsOnly := !strings.Contains(urlString, "?") && strings.Contains(urlString, "=")
	if isParamsOnly {
		query, err = url.ParseQuery(urlString)
		if err != nil {
			return "", errors.Wrapf(err, "wrong formatted url: %s", urlString)
		}
	} else {
		parsed, err = url.Parse(urlString)
		if err != nil {
			return "", errors.Wrapf(err, "wrong formatted url: %s", urlString)
		}

		query = parsed.Query()
	}

	excludeKeys = append(excludeKeys, defaultExcludeKeys...)
	hideKeys = append(hideKeys, defaultHideKeys...)

	for key, values := range query {
		if ArrayContains(hideKeys, strings.ToLower(key)) {
			query.Set(key, Hide(strings.Join(values, ",")))
		} else if ArrayContains(excludeKeys, strings.ToLower(key)) {
			query.Set(key, Mask(strings.Join(values, ",")))
		}
	}

	if isParamsOnly {
		return url.PathUnescape(query.Encode())
	}

	parsed.RawQuery = query.Encode()

	return MaskFQDN(parsed.String())
}

func MaskFQDN(rawURL string) (string, error) {
	const placeholder = "_password_"

	maskedURL := ""

	data, err := url.Parse(rawURL)
	if err == nil && rawURL != "" {
		_, ok := data.User.Password()
		if ok {
			user := data.User.Username()
			data.User = url.UserPassword(user, placeholder)
		} else {
			data.User = nil
		}

		data.RawQuery = data.Query().Encode()

		maskedURL = strings.Replace(data.String(), placeholder, "***", 1)
	}

	return maskedURL, err
}
