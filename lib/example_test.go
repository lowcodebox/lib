package app_lib

import (
	"fmt"
)

func ExampleFuncMapImpl_Timeparseany() {
	f := FuncMapImpl{}

	// Простой парсинг
	parsed, _ := f.Timeparseany("2024-04-04 11:11:11", false)
	fmt.Println(parsed)

	// Парсинг из привычного формата с зоной MSK
	parsed, _ = f.Timeparseany("04.04.2024 11:11:11 MSK", false)
	fmt.Println(parsed)

	// Выводим в UTC
	parsed, _ = f.Timeparseany("2024-04-04 11:11:11 MSK", true)
	fmt.Println(parsed)

	// Парсинг вместе с интервалом
	parsed, _ = f.Timeparseany("2024-04-04 11:11:11 MSK - 1d3h", false)
	fmt.Println(parsed)

	// Output:
	// 2024-04-04 11:11:11 +0000 UTC
	// 2024-04-04 11:11:11 +0300 MSK
	// 2024-04-04 08:11:11 +0000 UTC
	// 2024-04-03 08:11:11 +0300 MSK
}
