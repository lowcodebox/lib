package lib

import "strings"

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
