package lib

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"git.lowcodeplatform.net/packages/models"
)

func TestXServiceKey(t *testing.T) {
	cases := []struct {
		domain        string
		projectKey    []byte
		tokenInterval time.Duration
	}{
		{"algiva/orm", []byte("LKHlhb899Y09olUi"), 1000 * time.Second},
	}
	client := "testClient"

	for _, c := range cases {
		token, err := GenXServiceKey(c.domain, c.projectKey, c.tokenInterval, client)
		fmt.Println(token, err)

		//token = "e5vCdpG-s7Ya0hbt1vXE1pSrs8TAUbXjowBgcQSOYkaXlNf6wrxoea-QNEiLYVSu_fXz7jKEbLiODjghxwMzhw"
		//token = "-R7DGF0Q8zYpHxXzwk1xbAbTy5yJWmrGV5zQT0LIFTzOIm7CfsRfYLojo7G3WnABRMOR1VCh8XJcpJB0dRZM7m9FDDD63g4agyXdtWaeKj-mpHdWMzmhrQmBAkFH9euKjaoJ2dou5aj3TzI_mVCGBuTB4hWMABFF_RSO0J7wbGk6JJk84RRSkKAQMoxpN2TsN-YZ422G9DC6ICWzNloWUDH8-kPOwQSaKiJpWJ9vxrstyYrTiVCasGMaXjJfn5Sd"
		token = "GoJLcmb292x2BW41TZEr1lzqbZiASuye8Vi9IcgGBKgI_tXpAYsZPKTn81kPqnYa0PcEYc9PurkbBO1H45-Iz990rKNXxw9q2iSpQeNTbh4MU-y48AMLrjlse4_Yq3BUuTMPFt23p19ZldYYRDL_EJtu_Jpk7vRPojViLNp_CkyI5KC2hzZaBx2L4o_Fk0FnRHLzSEiAl7iE0IIfp6CYoGqTkDJ6Imcs1If0JUZP4GGYfKAmUj4czpdRQuBebl9xzPL4pmSGJrIGCKHS9gb4xU_lM2b69pKtJko5zAQzV7no74Hc8a7yA9toUjNAZxuhq15EeLMd1bZUBtDF6GPPSvJtuc-uD0EQae8B6vwLUwSZFx3AJi7Q2HA31kQ1Fp69dHjPfw11AgKeqocehTTHdZIrQ8A8Ds-YM4FTYg_tc9Eb-cJECRww7o1U2L9BdCZfy9IWMw05eNU2nBS41EtgQVXqwUqEdy2cWIqHslXqFsHUBp7tdhaTeok3AiPoOiZY8SFUskldl9iqWauLVtZd5Vdt7Ga-xGO6yfdeyGIHmUkcdiDPaafBOtTdCUsci7eFeqsnZRZ_bv19hV9GalcQc2VMF9EZqjgdV_Q-ZyOnPibGFMTWHc5l3RIanSLHtgauwXewAarpuejmffuhyLwcDJTe6XzntPZkg65PzCVLNZmGKnpcaadl7HpaThJQdCLiAkPaQv7cEYdBJZDwx3qHWQgZU5P_862okwTq7FA9dkXFezUJyGSjPOBFnyjfPzXe-3GByCAiNpvZj5wSUC1p9zIOlbojqxS0Xx5RNhoFpKaty0JkRVKubQYNMfGt6HAGN8prd-uSexKM8h8xJOfzexclDva8QMUbOk2doqE740dEWV_LOtkGctj_34_asmkC4bDrnUTiZ0wTGjaHFsFL0YyfE5QoWBpGTICQVJSrsJNXQwJYzDelJKSrSfb3-rfV4GTIJwlGYmz_Cr2gnPT0ZBZoV42tcvRbw8zA0_F7P2rKH3J3chSYvCGeCji3hnqHulOdVtgDZjTsEAbqbgUiRyPO8abmrN5uhygBys8ITP56Kf87ejH6g34aBjiAcRQYdjbbI_LO07CLZyo-USnMXcaStmn_PSp8i15nun1JAf0"

		// alg/org LKHlhb899Y09olUi -R7DGF0Q8zYpHxXzwk1xbAbTy5yJWmrGV5zQT0LIFTzOIm7CfsRfYLojo7G3WnABRMOR1VCh8XJcpJB0dRZM7m9FDDD63g4agyXdtWaeKj-mpHdWMzmhrQmBAkFH9euKjaoJ2dou5aj3TzI_mVCGBuTB4hWMABFF_RSO0J7wbGk6JJk84RRSkKAQMoxpN2TsN-YZ422G9DC6ICWzNloWUDH8-kPOwQSaKiJpWJ9vxrstyYrTiVCasGMaXjJfn5Sd -> true testClient

		valid, client2 := CheckXServiceKey(c.domain, c.projectKey, token)
		fmt.Println(c.domain, string(c.projectKey), token, "->", valid, client2)

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

func TestSetValidURI(t *testing.T) {
	const validUris = "pro/ru,lms/ru"
	var key = []byte("4160b6caea7ef66d")
	testCases := []struct {
		uri   string
		valid bool
	}{
		{"pro/ru", true},
		{"lms/ru", true},
		{"pro/ru/block/test", true},
		{"lms/pro/ru", false},
		{"lms/ruenus", false},
		{"pro/ru?key=val", true},
	}

	token, err := GenXServiceKey("lms/ru", key, time.Hour, "")
	if err != nil {
		t.Fatal(err)
	}

	token, err = SetValidURI(validUris, key, token)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		valid, err := IsValidURI(tc.uri, key, token)
		if err != nil {
			t.Fatal(err)
		}
		if valid != tc.valid {
			t.Errorf("ValidURI(%s) returned %t, want %t", tc.uri, valid, tc.valid)
		}
	}
}

func TestDecode(t *testing.T) {
	cases := []struct {
		domain     string
		projectKey []byte
		tokenType  string
		token      string
	}{
		{"algiva/orm", []byte("LKHlhb899Y09olUi"), "token", "-2okieyncmOOBckJM7CXaWZE5q6QUhlsmNL2LyrgTt74HgF_ecaLM_5UBmJBdS0g-ALdT2RfAGvH0DvX8IA78XLsOaA7UB4BFCyCBtFAOPs"},
		{"algiva/orm", []byte("LKHlhb899Y09olUi"), "X-Auth-Key", "GoJLcmb292x2BW41TZEr1lzqbZiASuye8Vi9IcgGBKgI_tXpAYsZPKTn81kPqnYa0PcEYc9PurkbBO1H45-Iz990rKNXxw9q2iSpQeNTbh4MU-y48AMLrjlse4_Yq3BUuTMPFt23p19ZldYYRDL_EJtu_Jpk7vRPojViLNp_CkyI5KC2hzZaBx2L4o_Fk0FnRHLzSEiAl7iE0IIfp6CYoGqTkDJ6Imcs1If0JUZP4GGYfKAmUj4czpdRQuBebl9xzPL4pmSGJrIGCKHS9gb4xU_lM2b69pKtJko5zAQzV7no74Hc8a7yA9toUjNAZxuhq15EeLMd1bZUBtDF6GPPSvJtuc-uD0EQae8B6vwLUwSZFx3AJi7Q2HA31kQ1Fp69dHjPfw11AgKeqocehTTHdZIrQ8A8Ds-YM4FTYg_tc9Eb-cJECRww7o1U2L9BdCZfy9IWMw05eNU2nBS41EtgQVXqwUqEdy2cWIqHslXqFsHUBp7tdhaTeok3AiPoOiZY8SFUskldl9iqWauLVtZd5Vdt7Ga-xGO6yfdeyGIHmUkcdiDPaafBOtTdCUsci7eFeqsnZRZ_bv19hV9GalcQc2VMF9EZqjgdV_Q-ZyOnPibGFMTWHc5l3RIanSLHtgauwXewAarpuejmffuhyLwcDJTe6XzntPZkg65PzCVLNZmGKnpcaadl7HpaThJQdCLiAkPaQv7cEYdBJZDwx3qHWQgZU5P_862okwTq7FA9dkXFezUJyGSjPOBFnyjfPzXe-3GByCAiNpvZj5wSUC1p9zIOlbojqxS0Xx5RNhoFpKaty0JkRVKubQYNMfGt6HAGN8prd-uSexKM8h8xJOfzexclDva8QMUbOk2doqE740dEWV_LOtkGctj_34_asmkC4bDrnUTiZ0wTGjaHFsFL0YyfE5QoWBpGTICQVJSrsJNXQwJYzDelJKSrSfb3-rfV4GTIJwlGYmz_Cr2gnPT0ZBZoV42tcvRbw8zA0_F7P2rKH3J3chSYvCGeCji3hnqHulOdVtgDZjTsEAbqbgUiRyPO8abmrN5uhygBys8ITP56Kf87ejH6g34aBjiAcRQYdjbbI_LO07CLZyo-USnMXcaStmn_PSp8i15nun1JAf0"},
	}

	for _, c := range cases {
		switch c.tokenType {
		case "X-Auth-Key":
			s, err := Decrypt(c.projectKey, c.token)
			if err != nil {
				t.Error(err)
			}
			tk := models.Token{}
			json.Unmarshal(&tk)

			token, err := decodeServiceKey(c.projectKey, tk.AccessKey)
			fmt.Println(fmt.Sprintf("%+v", token), s, err)
		case "token":
			token, err := decodeServiceKey(c.projectKey, c.domain)
			fmt.Println(fmt.Sprintf("%+v", token), c.domain, err)
		}
	}
}
