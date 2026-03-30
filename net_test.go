package lib

import (
	"fmt"
	"testing"
)

func TestGetPIDByPort(t *testing.T) {
	i := 8015

	pid, err := GetPIDByPort(i)
	if err != nil {
		fmt.Println(err)
	}

	println(pid)
}
