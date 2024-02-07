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
