package exec

import (
    "github.com/stretchr/testify/require"
    "testing"
)

func TestOption_Explain(t *testing.T) {
    type testCase struct {
        test string
        opt  *Option
        expectedResult string
    }

    testCases := []testCase{
        {
            test: "single char option",
            opt: &Option{
                Name: "r",
            },
            expectedResult: "-r",
        },
        {
            test: "multi char option",
            opt: &Option{
                Name: "run",
            },
            expectedResult: "--run",
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.test, func(t *testing.T) {
            require.Equal(t, testCase.expectedResult, testCase.opt.Explain())
        })
    }
}

func TestOption_String(t *testing.T) {
    type testCase struct {
        test string
        opt  *Option
        expectedPanic bool
    }

    testCases := []testCase{
        {
            test: "valid with value string type pointer",
            opt: &Option{
                Name:        "option",
                Type:        String,
                Value:       new(string),
            },
        },
        {
            test: "invalid value string type with no pointer",
            opt: &Option{
                Name:        "option",
                Type:        String,
                Value:       "string",
            },
            expectedPanic: true,
        },
        {
            test: "invalid value int type with no pointer",
            opt: &Option{
                Name:        "option",
                Type:        String,
                Value:       0,
            },
            expectedPanic: true,
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.test, func(t *testing.T) {
            if testCase.expectedPanic {
                require.Panics(t, func() {
                    _ = testCase.opt.String()
                })
            } else {
                require.Equal(t, *testCase.opt.Value.(*string), testCase.opt.String())
            }
        })
    }
}

func TestOption_Int(t *testing.T) {
    type testCase struct {
        test string
        opt  *Option
        expectedPanic bool
    }

    testCases := []testCase{
        {
            test: "valid with value int type pointer",
            opt: &Option{
                Name:        "option",
                Type:        Int,
                Value:       new(int),
            },
        },
        {
            test: "invalid value int type with no pointer",
            opt: &Option{
                Name:        "option",
                Type:        Int,
                Value:       0,
            },
            expectedPanic: true,
        },
        {
            test: "invalid value string type with no pointer",
            opt: &Option{
                Name:        "option",
                Type:        Int,
                Value:       "string",
            },
            expectedPanic: true,
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.test, func(t *testing.T) {
            if testCase.expectedPanic {
                require.Panics(t, func() {
                    testCase.opt.Int()
                })
            } else {
                require.Equal(t, *testCase.opt.Value.(*int), testCase.opt.Int())
            }
        })
    }
}

func TestOption_Bool(t *testing.T) {
    type testCase struct {
        test string
        opt  *Option
        expectedPanic bool
    }

    testCases := []testCase{
        {
            test: "valid with value bool type pointer",
            opt: &Option{
                Name:        "option",
                Type:        Bool,
                Value:       new(bool),
            },
        },
        {
            test: "invalid value bool type with no pointer",
            opt: &Option{
                Name:        "option",
                Type:        Bool,
                Value:       false,
            },
            expectedPanic: true,
        },
        {
            test: "invalid value string type with no pointer",
            opt: &Option{
                Name:        "option",
                Type:        Bool,
                Value:       "string",
            },
            expectedPanic: true,
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.test, func(t *testing.T) {
            if testCase.expectedPanic {
                require.Panics(t, func() {
                    testCase.opt.Bool()
                })
            } else {
                require.Equal(t, *testCase.opt.Value.(*bool), testCase.opt.Bool())
            }
        })
    }
}

func TestArgument_Explain(t *testing.T) {
    type testCase struct {
        test string
        arg  *Argument
        expectedResult string
    }

    testCases := []testCase{
        {
            test: "single arg",
            arg: &Argument{
                Name: "arg",
            },
            expectedResult: "<arg>",
        },
        {
            test: "multi arg",
            arg: &Argument{
                Name: "arg",
                Multiple: true,
            },
            expectedResult: "<arg>...",
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.test, func(t *testing.T) {
            require.Equal(t, testCase.expectedResult, testCase.arg.Explain())
        })
    }
}

func TestArgument_String(t *testing.T) {
    type testCase struct {
        test string
        arg  *Argument
        expectedPanic bool
    }

    testCases := []testCase{
        {
            test: "valid with value string type pointer",
            arg: &Argument{
                Name:        "argument",
                Type:        String,
                Value:       new(string),
            },
        },
        {
            test: "invalid value string type with no pointer",
            arg: &Argument{
                Name:        "argument",
                Type:        String,
                Value:       "string",
            },
            expectedPanic: true,
        },
        {
            test: "invalid value int type with no pointer",
            arg: &Argument{
                Name:        "argument",
                Type:        String,
                Value:       0,
            },
            expectedPanic: true,
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.test, func(t *testing.T) {
            if testCase.expectedPanic {
                require.Panics(t, func() {
                    _ = testCase.arg.String()
                })
            } else {
                require.Equal(t, *testCase.arg.Value.(*string), testCase.arg.String())
            }
        })
    }
}

func TestArgument_Int(t *testing.T) {
    type testCase struct {
        test string
        arg  *Argument
        expectedPanic bool
    }

    testCases := []testCase{
        {
            test: "valid with value int type pointer",
            arg: &Argument{
                Name:        "argument",
                Type:        Int,
                Value:       new(int),
            },
        },
        {
            test: "invalid value int type with no pointer",
            arg: &Argument{
                Name:        "argument",
                Type:        Int,
                Value:       0,
            },
            expectedPanic: true,
        },
        {
            test: "invalid value string type with no pointer",
            arg: &Argument{
                Name:        "argument",
                Type:        Int,
                Value:       "string",
            },
            expectedPanic: true,
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.test, func(t *testing.T) {
            if testCase.expectedPanic {
                require.Panics(t, func() {
                    testCase.arg.Int()
                })
            } else {
                require.Equal(t, *testCase.arg.Value.(*int), testCase.arg.Int())
            }
        })
    }
}

func TestArgument_Bool(t *testing.T) {
    type testCase struct {
        test string
        arg  *Argument
        expectedPanic bool
    }

    testCases := []testCase{
        {
            test: "valid with value bool type pointer",
            arg: &Argument{
                Name:        "argument",
                Type:        Bool,
                Value:       new(bool),
            },
        },
        {
            test: "invalid value bool type with no pointer",
            arg: &Argument{
                Name:        "argument",
                Type:        Bool,
                Value:       false,
            },
            expectedPanic: true,
        },
        {
            test: "invalid value string type with no pointer",
            arg: &Argument{
                Name:        "argument",
                Type:        Bool,
                Value:       "string",
            },
            expectedPanic: true,
        },
    }

    for _, testCase := range testCases {
        t.Run(testCase.test, func(t *testing.T) {
            if testCase.expectedPanic {
                require.Panics(t, func() {
                    testCase.arg.Bool()
                })
            } else {
                require.Equal(t, *testCase.arg.Value.(*bool), testCase.arg.Bool())
            }
        })
    }
}
