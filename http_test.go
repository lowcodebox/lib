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
		{"/lms/ru", []byte("LKHlhb899Y09olUi"), 10 * time.Second},
	}

	for _, c := range cases {
		token, err := GenXServiceKey(c.domain, c.projectKey, c.tokenInterval)
		fmt.Println(token, err)

		valid := CheckXServiceKey(c.domain, c.projectKey, token)
		fmt.Println(valid)

	}
}
