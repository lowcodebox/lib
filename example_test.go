package lib

import (
	"fmt"
)

func ExampleTimeParse() {
	// Простой парсинг
	res, _ := TimeParse("2024-04-04 11:11:11", false)
	fmt.Println(res)

	// Парсинг из привычного формата с зоной MSK
	res, _ = TimeParse("04.04.2024 11:11:11 MSK", false)
	fmt.Println(res)

	// Выводим в UTC
	res, _ = TimeParse("2024-04-04 11:11:11 MSK", true)
	fmt.Println(res)

	// Парсинг вместе с интервалом
	res, _ = TimeParse("2024-04-04 11:11:11 MSK - 1d3h", false)
	fmt.Println(res)

	// Парсинг с несколькими интервалами
	res, _ = TimeParse("2024-04-04 11:11:11 MSK - 1d - 3h", false)
	fmt.Println(res)

	// Output:
	// 2024-04-04 11:11:11 +0000 UTC
	// 2024-04-04 11:11:11 +0300 MSK
	// 2024-04-04 08:11:11 +0000 UTC
	// 2024-04-03 08:11:11 +0300 MSK
	// 2024-04-03 08:11:11 +0300 MSK
}
