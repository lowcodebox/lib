package lib

import "strings"

// HideExceptFirstAndLast экранирует в строке символы между указанным количеством
// opt0 - сколько символов оставить сначала строки (по-умолчанию 3)
// opt1 - сколько символов оставить в конце строки (по-умолчанию 3)
func HideExceptFirstAndLast(str string, opt ...int) string {
	first, last := 1, 1
	if len(opt) > 0 {
		first = opt[0]
	}
	if len(opt) > 1 {
		last = opt[1]
	}

	if len(str) <= first+last {
		return str
	}

	hidden := str[:first] + strings.Repeat("*", len(str)-first-last) + str[len(str)-last:]
	if len(hidden) > 15 {
		return hidden[:15]
	}
	return hidden
}
