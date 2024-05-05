package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

//func test() {
//	fmt.Println("\nCARDS: 546901XXXXXX6304")
//	fmt.Println("SELECT * from cards where card_mask = '546901XXXXXX6304'\n")
//	fmt.Println("CARD (OLD): 546901XXXXXX6304")
//
//	fmt.Println(supplementTime(0), " - card (old): added_time" )
//	fmt.Println(supplementTime(1628754109), " - card: last_payment_time\n")
//
//	fmt.Println("CARD (NEW): 546901XXXXXX6304")
//	fmt.Println(supplementTime(1630041399), " - card: added_time", )
//	fmt.Println(supplementTime(1629913153), " - card: last_payment_time\n")
//
//	fmt.Println("PAYMENTS")
//	fmt.Println("SELECT * from payments where card_mask = '546901XXXXXX6304'")
//	fmt.Println(supplementTime(1625734996), " - payment: bill_time")
//	fmt.Println(supplementTime(1628752541), " - payment: bill_time")
//	fmt.Println(supplementTime(1628319338), " - payment: bill_time")
//
//
//}

func supplementTime(t int64) (res interface{}) {
	if t == 0 {
		return nil
	}
	for ; t < 1000000000000000000; t *= 10 {
	}
	res = time.Unix(0, t)

	return res
}

func testpostgres() {
	host := "127.0.0.1"
	port := "5432"
	user := "postgres"
	password := "password"
	dbname := "postgres"

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	result, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println("Tidak Konek DB Errornya :", err)
	}

	//cmdTag, err := result.Exec("INSERT INTO test (id) values ($1)", "3")
	cmdTag, err := result.Exec("UPDATE test set bin = $1", supplementTime(1212321312312312))

	if err != nil {
		fmt.Println("can't save card to db:", err)
	}

	fmt.Println(cmdTag)
}

func main1() {
	strv := `
{"statusCode":404,"error":"Not Found","message":"Not Found"}`

	fmt.Println(validateBodyType([]byte(strv), "text"))
}

// проверка на валидность типа данных в ответе
// проверка на типы (expectedType): html, json, text
func validateBodyType(strBody []byte, expectedType string) bool {
	etype := http.DetectContentType(strBody)

	switch expectedType {
	case "html":
		if etype == "text/html; charset=utf-8" {
			return true
		}
	case "text":
		if strings.Contains(etype, "text") {
			return true
		}
	case "json":
		return json.Valid(strBody)
	}

	return false
}
