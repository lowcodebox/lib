package lib

import (
	"strconv"
	"strings"
)

// HideExceptFirstAndLast экранирует в строке символы между указанным количеством
// opt0 - сколько символов оставить сначала строки (по-умолчанию 3)
// opt1 - сколько символов оставить в конце строки (по-умолчанию 3)
func HideExceptFirstAndLast(str string, opt ...int) string {
	prefCount := 1
	postCount := 1

	if len(opt) > 0 {
		prefCount = opt[0]
	}
	if len(opt) > 1 {
		postCount = opt[1]
	}

	lengthOfPan := len(str)
	builder := strings.Builder{}

	for i, n := range str {
		switch {
		case i < prefCount:
			builder.WriteRune(n)
		case i >= lengthOfPan-postCount:
			builder.WriteRune(n)
		default:
			builder.WriteString("*")
		}
	}

	if builder.Len() > 16 {
		return builder.String()[:15]
	}
	return builder.String()
}

// FirstVal возвращает первое непустое значение
func FirstVal[T comparable](vals ...T) T {
	var null T
	for _, val := range vals {
		if val != null {
			return val
		}
	}

	return null
}

func ParseInt(s string) (i int, ok bool) {
	n, err := strconv.Atoi(s)

	return n, err == nil
}

func ParseInt64(s string) (i int64, ok bool) {
	n, err := strconv.ParseInt(s, 10, 64)

	return n, err == nil
}

func ParseFloat(s string) (i float64, ok bool) {
	n, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)

	return n, err == nil
}

func ArrayDelete[T any](slice []T, i int) []T {
	return append(slice[:i], slice[i+1:]...)
}
