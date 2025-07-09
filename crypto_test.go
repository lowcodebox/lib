package lib

import (
	"fmt"
	"testing"
	"time"
)

func TestXServiceKey(t *testing.T) {
	t.Parallel()
	cases := []struct {
		domain        string
		projectKey    []byte
		tokenInterval time.Duration
	}{
		{"algiva/orm", []byte("LKHlhb899Y09olUi"), 1000 * time.Second},
	}
	client := "testClient"

	for _, c := range cases {
		token, err := GenXServiceKey(c.domain, c.projectKey, c.tokenInterval, client)
		fmt.Println(token, err)

		token = "e5vCdpG-s7Ya0hbt1vXE1pSrs8TAUbXjowBgcQSOYkaXlNf6wrxoea-QNEiLYVSu_fXz7jKEbLiODjghxwMzhw"
		valid, client2 := CheckXServiceKey(c.domain, c.projectKey, token)
		fmt.Println(c.domain, string(c.projectKey), token, "->", valid, client2)

	}
}

func TestEncryptArgon2(t *testing.T) {
	t.Parallel()
	cases := []struct {
		domain        string
		projectKey    []byte
		tokenInterval time.Duration
	}{
		{"algiva/orm", []byte("LKHlhb899Y09olUi"), 1000 * time.Second},
	}

	for _, c := range cases {
		token, err := EncryptArgon2(c.domain, nil)
		if err != nil {
			t.Errorf("error %s", err)
		}

		boolRes := CheckArgon2(c.domain, token)
		if !boolRes {
			t.Errorf("Result was incorrect, got: %t, want: %t.", true, false)
		}

		boolRes = CheckArgon2(c.domain+"randtext", token)
		if boolRes {
			t.Errorf("Result was incorrect, got: %t, want: %t.", true, false)
		}
	}
}

func TestSetValidURI(t *testing.T) {
	t.Parallel()
	const validUris = "pro/ru,lms/ru"
	var key = []byte("4160b6caea7ef66d")
	testCases := []struct {
		uri   string
		valid bool
	}{
		{"pro/ru", true},
		{"lms/ru", true},
		{"pro/ru/block/test", true},
		{"lms/pro/ru", false},
		{"lms/ruenus", false},
		{"pro/ru?key=val", true},
	}

	token, err := GenXServiceKey("lms/ru", key, time.Hour, "")
	if err != nil {
		t.Fatal(err)
	}

	token, err = SetValidURI(validUris, key, token)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		valid, err := IsValidURI(tc.uri, key, token)
		if err != nil {
			t.Fatal(err)
		}
		if valid != tc.valid {
			t.Errorf("ValidURI(%s) returned %t, want %t", tc.uri, valid, tc.valid)
		}
	}
}
