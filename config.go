package exec

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type config struct {
	Name  string
	value interface{}
}

func (c *config) Value() interface{} {
	v := reflect.ValueOf(c.value)

	if v.Kind() == reflect.Func {
		t := v.Type()

		if t.NumIn() != 0 && t.NumOut() != 1 {
			panic("Function type must have no input parameters and a single return value")
		}

		if t.Out(0).Kind().String() != "interface" {
			panic("Function return value must be an interface{}")
		}

		c.value = v.Call(nil)[0].Interface()
	}
	return c.value
}

func (c *config) String() string {
	return strings.TrimSpace(fmt.Sprint(c.Value()))
}

func (c *config) Int() int {
	return c.Value().(int)
}

func (c *config) Int64() int64 {
	return c.Value().(int64)
}

func (c *config) Bool() bool {
	return c.Value().(bool)
}

func (c *config) Slice() []string {
	return c.Value().([]string)
}

func (c *config) Time() time.Time {
	return c.Value().(time.Time)
}
