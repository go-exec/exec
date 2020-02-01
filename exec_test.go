package exec_test

import (
	"github.com/go-exec/exec"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewArgument(t *testing.T) {
	arg := &exec.Argument{
		Name:        "name",
		Type:        0,
		Default:     nil,
		Multiple:    false,
		Description: "description",
		Value:       nil,
	}
	require.Equal(t, exec.NewArgument(arg.Name, arg.Description), arg)
}

func TestAddArgument(t *testing.T) {
	arg := &exec.Argument{
		Name:        "test",
		Type:        0,
		Default:     nil,
		Multiple:    false,
		Description: "",
		Value:       nil,
	}
	exec.AddArgument(arg)

	require.Equal(t, arg, exec.Arguments[arg.Name])
}

func TestGetArgument(t *testing.T) {
	type testCase struct {
		test string
		name string
		arg  *exec.Argument
	}

	testCases := []testCase{
		{
			test: "valid argument",
			name: "valid",
			arg: &exec.Argument{
				Name: "valid",
			},
		},
		{
			test: "invalid argument",
			name: "invalid",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			if testCase.arg != nil {
				exec.AddArgument(testCase.arg)
			}

			require.Equal(t, exec.GetArgument(testCase.name), testCase.arg)
		})
	}
}
