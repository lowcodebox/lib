package app_lib

import (
	"fmt"
)

func ExampleFuncMapImpl_Timeparseany() {
	f := FuncMapImpl{}

	// Простой парсинг
	parsed := f.Timeparseany("2024-04-04 11:11:11", false)
	fmt.Println(parsed.Time)

	// Парсинг из привычного формата с зоной MSK
	parsed = f.Timeparseany("04.04.2024 11:11:11 MSK", false)
	fmt.Println(parsed.Time)

	// Выводим в UTC
	parsed = f.Timeparseany("2024-04-04 11:11:11 MSK", true)
	fmt.Println(parsed.Time)

	// Парсинг вместе с интервалом
	parsed = f.Timeparseany("2024-04-04 11:11:11 MSK - 1d3h", false)
	fmt.Println(parsed.Time)

	// Парсинг с несколькими интервалами
	parsed = f.Timeparseany("2024-04-04 11:11:11 MSK - 1d - 3h", false)
	fmt.Println(parsed.Time)

	// Output:
	// 2024-04-04 11:11:11 +0000 UTC
	// 2024-04-04 11:11:11 +0300 MSK
	// 2024-04-04 08:11:11 +0000 UTC
	// 2024-04-03 08:11:11 +0300 MSK
	// 2024-04-03 08:11:11 +0300 MSK
}
