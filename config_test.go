package exec

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func TestConfig_Value(t *testing.T) {
	type testCase struct {
		test          string
		cfg           *config
		expectedPanic interface{}
	}

	testCases := []testCase{
		{
			test: "valid simple value",
			cfg: &config{
				value: "",
			},
		},
		{
			test: "valid func value",
			cfg: &config{
				value: func() interface{} {
					return "value"
				},
			},
		},
		{
			test: "panics on invalid func with input param",
			cfg: &config{
				value: func(s string) interface{} {
					return "value"
				},
			},
			expectedPanic: "Function type must have no input parameters",
		},
		{
			test: "panics on invalid func with more than one return param",
			cfg: &config{
				value: func() (interface{}, string) {
					return "value", "one"
				},
			},
			expectedPanic: "Function type must have a single return value",
		},
		{
			test: "panics on invalid func with no input param",
			cfg: &config{
				value: func() string {
					return "value"
				},
			},
			expectedPanic: "Function return value must be an interface{}",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			if testCase.expectedPanic != nil {
				require.PanicsWithValue(t, testCase.expectedPanic, func() {
					testCase.cfg.Value()
				})
			} else {
				require.Equal(t, testCase.cfg.value, testCase.cfg.Value())
			}
		})
	}
}

func TestConfig_String(t *testing.T) {
	cfg := &config{
		Name:  "name",
		value: "string",
	}

	require.Equal(t, cfg.value, cfg.String())
	require.IsType(t, reflect.TypeOf(cfg.value), reflect.TypeOf(cfg.String()))
}

func TestConfig_Int(t *testing.T) {
	cfg := &config{
		Name:  "name",
		value: 1,
	}

	require.Equal(t, cfg.value, cfg.Int())
	require.IsType(t, reflect.TypeOf(cfg.value), reflect.TypeOf(cfg.Int()))
}

func TestConfig_Int64(t *testing.T) {
	cfg := &config{
		Name:  "name",
		value: int64(1),
	}

	require.Equal(t, cfg.value, cfg.Int64())
	require.IsType(t, reflect.TypeOf(cfg.value), reflect.TypeOf(cfg.Int64()))
}

func TestConfig_Bool(t *testing.T) {
	cfg := &config{
		Name:  "name",
		value: true,
	}

	require.Equal(t, cfg.value, cfg.Bool())
	require.IsType(t, reflect.TypeOf(cfg.value), reflect.TypeOf(cfg.Bool()))
}

func TestConfig_Slice(t *testing.T) {
	cfg := &config{
		Name:  "name",
		value: []string{"a", "b"},
	}

	require.Equal(t, cfg.value, cfg.Slice())
	require.IsType(t, reflect.TypeOf(cfg.value), reflect.TypeOf(cfg.Slice()))
}

func TestConfig_Time(t *testing.T) {
	cfg := &config{
		Name:  "name",
		value: time.Now(),
	}

	require.Equal(t, cfg.value, cfg.Time())
	require.IsType(t, reflect.TypeOf(cfg.value), reflect.TypeOf(cfg.Time()))
}
