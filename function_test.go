package lib

import (
	"fmt"
	"testing"
)

func TestSearchConfig(t *testing.T) {
	t.Skip()

	cases := []struct {
		path     string
		configfile 	 string
	}{
		{"/Users/ivan/go/src/git.lowcodeplatform.net/buildbox/upload/buildbox/bin/proxy/darwin/v1.2.0", "2021-04-01T09-32-39Z-515f56"},
	}

	for _, c := range cases {
		res, err := SearchConfig(c.configfile, c.path)
		fmt.Println(res, err)
	}
}