package lib

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFilesToMap(t *testing.T) {
	res, err := ReadFilesToMap("./pkg", true)
	assert.NoError(t, err)

	for k, v := range res {
		fmt.Println(k, len(v))
	}
}
