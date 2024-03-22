package lib

import (
	"fmt"
	"testing"
	"time"
)

func TestXServiceKey(t *testing.T) {
	cases := []struct {
		domain        string
		projectKey    []byte
		tokenInterval time.Duration
	}{
		{"algiva/orm", []byte("LKHlhb899Y09olUi"), 1000 * time.Second},
	}

	for _, c := range cases {
		token, err := GenXServiceKey(c.domain, c.projectKey, c.tokenInterval)
		fmt.Println(token, err)

		token = "e5vCdpG-s7Ya0hbt1vXE1pSrs8TAUbXjowBgcQSOYkaXlNf6wrxoea-QNEiLYVSu_fXz7jKEbLiODjghxwMzhw"
		valid := CheckXServiceKey(c.domain, c.projectKey, token)
		fmt.Println(c.domain, string(c.projectKey), token, "->", valid)

	}
}

func TestEncryptArgon2(t *testing.T) {
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
