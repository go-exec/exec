package exec

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestOutput_HasError(t *testing.T) {
	o := &Output{
		err: errors.New(""),
	}

	require.True(t, o.HasError())
}

func TestOutput_String(t *testing.T) {
	o := &Output{}

	require.True(t, o.String() == o.text)
}

func TestOutput_Int(t *testing.T) {
	type testCase struct {
		test     string
		output   *Output
		expected int
	}

	testCases := []testCase{
		{
			test: "valid string to int via atoi",
			output: &Output{
				text: "1",
			},
			expected: 1,
		},
		{
			test: "invalid string to int via atoi",
			output: &Output{
				text: "string",
			},
			expected: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			require.Equal(t, testCase.expected, testCase.output.Int())
		})
	}
}

func TestOutput_Bool(t *testing.T) {
	type testCase struct {
		test     string
		output   *Output
		expected bool
	}

	testCases := []testCase{
		{
			test: "valid true string to bool",
			output: &Output{
				text: "true",
			},
			expected: true,
		},
		{
			test: "valid !true string to bool",
			output: &Output{
				text: "string",
			},
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			require.Equal(t, testCase.expected, testCase.output.Bool())
		})
	}
}

func TestOutput_Slice(t *testing.T) {
	type testCase struct {
		test      string
		output    *Output
		separator string
		expected  []string
	}

	testCases := []testCase{
		{
			test: "valid splice a,b,c",
			output: &Output{
				text: "a,b,c",
			},
			separator: ",",
			expected:  []string{"a", "b", "c"},
		},
		{
			test: "valid splice empty",
			output: &Output{
				text: "",
			},
			separator: ",",
			expected:  []string{""},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			require.Equal(t, testCase.expected, testCase.output.Slice(testCase.separator))
		})
	}
}
