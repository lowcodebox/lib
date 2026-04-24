package lib

import (
	"fmt"
	"testing"
)

func TestGetPIDByPort(t *testing.T) {
	i := 80

	pid, err := GetPIDByPort(i)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(pid)
}
