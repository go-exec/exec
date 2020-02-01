package exec

import (
	"strconv"
	"strings"
)

type output struct {
	text string
	err  error
}

func (o output) HasError() bool {
	return o.err != nil
}

func (o output) String() string {
	return o.text
}

func (o output) Int() int {
	i, err := strconv.Atoi(o.text)
	if err == nil {
		return i
	}
	return 0
}

func (o output) Bool() bool {
	return "true" == o.text
}

func (o output) Slice(sep string) []string {
	return strings.Split(o.text, sep)
}
