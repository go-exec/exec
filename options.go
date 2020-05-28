package exec

import (
	"fmt"
)

type valType int

const (
	// String value type
	String valType = iota
	// Bool value type
	Bool
	// Int value type
	Int
)

// Option represents an optional flag.
type Option struct {
	Name        string
	Type        valType
	Default     interface{}
	Description string
	Value       interface{}
}

// Explain prepends the option name with one or two leading dashes and returns
// it. It is used to generate help texts.
func (opt Option) Explain() string {
	prefix := "-"
	if len(opt.Name) > 1 {
		prefix = "--"
	}
	return prefix + opt.Name
}

// String casts a value to a string and panics on failure.
func (opt Option) String() string {
	return *opt.Value.(*string)
}

// Bool casts a value to a bool and panics on failure.
func (opt Option) Bool() bool {
	return *opt.Value.(*bool)
}

// Int casts a value to an int and panics on failure.
func (opt Option) Int() int {
	return *opt.Value.(*int)
}

// Argument represents a required argument.
type Argument struct {
	Name        string
	Type        valType
	Default     interface{}
	Multiple    bool
	Description string
	Value       interface{}
	sequence    int
}

// Explain places the argument name in angle brackets and appends three dots to
// it in order to indicate multiple arguments. It is used to generate help
// texts.
func (arg Argument) Explain() string {
	format := "<%s>"
	if arg.Multiple {
		format += "..."
	}
	return fmt.Sprintf(format, arg.Name)
}

// String casts a value to a string and panics on failure.
func (arg Argument) String() string {
	return *arg.Value.(*string)
}

// Bool casts a value to a bool and panics on failure.
func (arg Argument) Bool() bool {
	return *arg.Value.(*bool)
}

// Int casts a value to an int and panics on failure.
func (arg Argument) Int() int {
	return *arg.Value.(*int)
}

type sortArguments []*Argument

func (slice sortArguments) Len() int {
	return len(slice)
}

func (slice sortArguments) Less(i, j int) bool {
	return slice[i].sequence < slice[j].sequence
}

func (slice sortArguments) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
