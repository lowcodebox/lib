package lib

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSearchConfig(t *testing.T) {
	t.Skip()

	cases := []struct {
		path       string
		configfile string
	}{
		{"/Users/ivan/go/src/git.lowcodeplatform.net/buildbox/upload/buildbox/bin/proxy/darwin/v1.2.0", "2021-04-01T09-32-39Z-515f56"},
	}

	for _, c := range cases {
		res, err := SearchConfig(c.configfile, c.path)
		fmt.Println(res, err)
	}
}

func TestTimeParse(t *testing.T) {
	res, err := TimeParse("04.04.2024 11:11:11 MSK - 1d3h", false)
	assert.Nil(t, err, "parsing time")
	exp := time.Date(2024, 4, 3, 8, 11, 11, 0, time.FixedZone("Europe/Moscow", 3*3600)).Local()
	assert.Equal(t, exp, res, "check result")
	fmt.Println(res)

	res, err = TimeParse("04.04.2024 11:11:11 MSK - 1d3h", true)
	assert.Nil(t, err, "parsing time")
	exp = time.Date(2024, 4, 3, 8, 11, 11, 0, time.FixedZone("Europe/Moscow", 3*3600)).UTC()
	assert.Equal(t, exp, res, "check result")
	fmt.Println(res)

	res, err = TimeParse("04.04.2024 11:11:11 MSK - 1d - 3h", true)
	assert.Nil(t, err, "parsing time")
	exp = time.Date(2024, 4, 3, 8, 11, 11, 0, time.FixedZone("Europe/Moscow", 3*3600)).UTC()
	assert.Equal(t, exp, res, "check result")
	fmt.Println(res)
}
