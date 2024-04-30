package types

import (
	"fmt"
	"strings"
)

const (
	longLenPan  = 16
	shortLenPan = 12
)

var defaultExcludeKeys = []string{
	"email",
	"card",
	"first_name",
	"firstname",
	"last_name",
	"lastname",
	"cvc",
	"cvc2",
	"csc",
	"csc2",
	"pan",
	"$pan",
	"cardnumber",
	"$cvc",
	"$cvc2",
	"pg_card_pan",
	"pg_card_cvc",
	"hpan",
}

var defaultHideKeys = []string{
	"password",
	"secret",
	"client_secret",
	"access_token",
}

func Mask(value string) string {
	switch {
	case len(value) > longLenPan:
		return value[:6] + fmt.Sprintf("---(%d)---", len(value)-longLenPan) + value[len(value)-4:]
	case len(value) > shortLenPan:
		return value[:4] + strings.Repeat("-", len(value)-shortLenPan) + value[len(value)-4:]
	default:
		return Hide(value)
	}
}

func Hide(value string) string {
	return strings.Repeat("-", len(value))
}
