package exec

import (
	"strconv"
	"strings"
)

type Output struct {
	text string
	err  error
}

func (o Output) HasError() bool {
	return o.err != nil
}

func (o Output) String() string {
	return o.text
}

func (o Output) Int() int {
	i, err := strconv.Atoi(o.text)
	if err == nil {
		return i
	}
	return 0
}

func (o Output) Bool() bool {
	return "true" == o.text
}

func (o Output) Slice(sep string) []string {
	return strings.Split(o.text, sep)
}
