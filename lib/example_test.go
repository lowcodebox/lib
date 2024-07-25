package app_lib

import (
	"fmt"
)

func ExampleFuncImpl_Timeparseany() {
	NewFuncMap(nil, nil, nil, "", nil, nil)

	// Простой парсинг
	res := Funcs.Timeparseany("2024-04-04 11:11:11", false)
	fmt.Println(res.Time)

	// Парсинг из привычного формата с зоной MSK
	res = Funcs.Timeparseany("04.04.2024 11:11:11 MSK", false)
	fmt.Println(res.Time)

	// Выводим в UTC
	res = Funcs.Timeparseany("2024-04-04 11:11:11 UTC+1", true)
	fmt.Println(res.Time)

	// Парсинг вместе с интервалом
	res = Funcs.Timeparseany("2024-04-04 11:11:11 MSK - 1d3h", false)
	fmt.Println(res.Time)

	// Парсинг с несколькими интервалами
	res = Funcs.Timeparseany("2024-04-04 11:11:11 MSK - 1d - 3h", false)
	fmt.Println(res.Time)

	// Output:
	// 2024-04-04 11:11:11 +0000 UTC
	// 2024-04-04 11:11:11 +0300 MSK
	// 2024-04-04 10:11:11 +0000 UTC
	// 2024-04-03 08:11:11 +0300 MSK
	// 2024-04-03 08:11:11 +0300 MSK
}
