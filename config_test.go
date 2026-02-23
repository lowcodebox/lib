package lib

import (
	"fmt"
	"testing"

	"git.lowcodeplatform.net/packages/models"
)

type Config struct {
	models.Config
}

func TestConfigLoad(t *testing.T) {
	var cfg Config
	cases := []struct {
		config string
	}{
		{"./pkg/tests/config"},
	}

	for _, c := range cases {
		_, err := ConfigLoad(c.config, &cfg)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(cfg, err)
	}
}
