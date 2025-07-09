package lib_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	lib "git.edtech.vm.prod-6.cloud.el/fabric/lib"
)

func TestHideExceptFirstAndLast_Default(t *testing.T) {
	t.Parallel()
	type tc struct {
		input string
		want  string
	}
	cases := []tc{
		{"abcdef", "a****f"}, // default 1,1
		{"a", "a"},           // single char
		{"ab", "ab"},         // two chars
		{"abc", "a*c"},       // three chars
		{"abcd", "a**d"},     // four chars
	}
	for _, c := range cases {
		got := lib.HideExceptFirstAndLast(c.input)
		assert.Equal(t, c.want, got, "input=%q", c.input)
	}
}

func TestHideExceptFirstAndLast_WithOptions(t *testing.T) {
	t.Parallel()
	type tc struct {
		input      string
		pref, post int
		want       string
	}
	cases := []tc{
		{"abcdef", 2, 2, "ab**ef"},
		{"abcdef", 3, 1, "abc**f"}, // avoid compile error for blank else
		{"abcdef", 1, 3, "a**def"},
		{"abcdef", 2, 4, "abcdef"}, // post > len-input => acts like default for overlap
	}
	for idx, c := range cases {
		t.Run(fmt.Sprintf("Test%d", idx), func(t *testing.T) {
			got := lib.HideExceptFirstAndLast(c.input, c.pref, c.post)
			assert.Equal(t, c.want, got,
				"input=%q, pref=%d, post=%d", c.input, c.pref, c.post)
		})
	}
}

func TestHideExceptFirstAndLast_Truncate(t *testing.T) {
	t.Parallel()
	// create a string longer than 16 chars
	input := strings.Repeat("X", 20)
	// with default opt, builder produces "X" + 18*"*" + "X" = len 20, then truncated to 15
	got := lib.HideExceptFirstAndLast(input)
	assert.Len(t, got, 15)
	// should start with "X" and end with "*"
	assert.Equal(t, "X", string(got[0]))
	assert.Equal(t, "*", string(got[len(got)-1]))
}

func TestFirstVal(t *testing.T) {
	t.Parallel()
	// int
	i := lib.FirstVal(0, 0, 5, 7)
	assert.Equal(t, 5, i)
	// string
	s := lib.FirstVal("", "first", "second")
	assert.Equal(t, "first", s)
	// all zero => returns zero value
	s2 := lib.FirstVal("")
	assert.Equal(t, "", s2)
}

func TestParseInt(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  int
		ok    bool
	}{
		{"123", 123, true},
		{"0", 0, true},
		{"-5", -5, true},
		{"abc", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := lib.ParseInt(c.input)
		assert.Equal(t, c.ok, ok, "input=%q ok", c.input)
		assert.Equal(t, c.want, got, "input=%q value", c.input)
	}
}

func TestParseInt64(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  int64
		ok    bool
	}{
		{"1234567890123", 1234567890123, true},
		{"0", 0, true},
		{"-42", -42, true},
		{"abc", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := lib.ParseInt64(c.input)
		assert.Equal(t, c.ok, ok, "input=%q ok", c.input)
		assert.Equal(t, c.want, got, "input=%q value", c.input)
	}
}

func TestParseFloat(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  float64
		ok    bool
	}{
		{"1.23", 1.23, true},
		{"4,56", 4.56, true},
		{"0", 0, true},
		{"-7.5", -7.5, true},
		{"abc", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := lib.ParseFloat(c.input)
		assert.Equal(t, c.ok, ok, "input=%q ok", c.input)
		if ok {
			assert.InDelta(t, c.want, got, 1e-9, "input=%q value", c.input)
		}
	}
}

func TestArrayDelete(t *testing.T) {
	t.Parallel()
	type intCase struct {
		name  string
		slice []int
		idx   int
		want  []int
	}
	intCases := []intCase{
		{"delete at start", []int{1, 2, 3, 4, 5}, 0, []int{2, 3, 4, 5}},
		{"delete in middle", []int{1, 2, 3, 4, 5}, 2, []int{1, 2, 4, 5}},
		{"delete at end", []int{1, 2, 3, 4, 5}, 4, []int{1, 2, 3, 4}},
		{"single element", []int{42}, 0, []int{}},
	}
	for _, tc := range intCases {
		t.Run(tc.name, func(t *testing.T) {
			got := lib.ArrayDelete(tc.slice, tc.idx)
			assert.Equal(t, tc.want, got)
		})
	}

	type strCase struct {
		name  string
		slice []string
		idx   int
		want  []string
	}
	strCases := []strCase{
		{"delete middle", []string{"a", "b", "c"}, 1, []string{"a", "c"}},
		{"delete first", []string{"x", "y"}, 0, []string{"y"}},
		{"delete last", []string{"x", "y"}, 1, []string{"x"}},
		{"single element", []string{"only"}, 0, []string{}},
	}
	for _, tc := range strCases {
		t.Run(tc.name, func(t *testing.T) {
			got := lib.ArrayDelete(tc.slice, tc.idx)
			assert.Equal(t, tc.want, got)
		})
	}
}
