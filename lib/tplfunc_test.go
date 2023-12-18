package app_lib

import (
	"testing"
)

func Test_csvtosliсemap(t *testing.T) {
	in := "field1,field2\n2,3"

	NewFuncMap(nil, nil)
	res, err := Funcs.csvtosliсemap([]byte(in))
	if err != nil {
		t.Errorf("Should not produce an error")
	}

	if res[0]["field1"] != "2" {
		t.Errorf("Result was incorrect, got: %s, want: %s.", res[0]["field1"], "2")
	}
}
