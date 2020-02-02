package exec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewArgument(t *testing.T) {
	arg := &Argument{
		Name:        "name",
		Type:        0,
		Default:     nil,
		Multiple:    false,
		Description: "description",
		Value:       nil,
	}
	require.Equal(t, NewArgument(arg.Name, arg.Description), arg)
}

func TestAddArgument(t *testing.T) {
	arg := &Argument{
		Name:        "test",
		Type:        0,
		Default:     nil,
		Multiple:    false,
		Description: "",
		Value:       nil,
	}
	AddArgument(arg)

	require.Equal(t, arg, Arguments[arg.Name])
}

func TestGetArgument(t *testing.T) {
	type testCase struct {
		test string
		name string
		arg  *Argument
	}

	testCases := []testCase{
		{
			test: "valid argument",
			name: "valid",
			arg: &Argument{
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
				AddArgument(testCase.arg)
			}

			require.Equal(t, GetArgument(testCase.name), testCase.arg)
		})
	}
}

func TestNewOption(t *testing.T) {
	opt := &Option{
		Name:        "name",
		Type:        0,
		Default:     nil,
		Description: "description",
		Value:       nil,
	}
	require.Equal(t, NewOption(opt.Name, opt.Description), opt)
}

func TestAddOption(t *testing.T) {
	opt := &Option{
		Name:        "name",
		Type:        0,
		Default:     nil,
		Description: "description",
		Value:       nil,
	}
	AddOption(opt)

	require.Equal(t, opt, Options[opt.Name])
}

func TestGetOption(t *testing.T) {
	type testCase struct {
		test string
		name string
		opt  *Option
	}

	testCases := []testCase{
		{
			test: "valid option",
			name: "valid",
			opt: &Option{
				Name: "valid",
			},
		},
		{
			test: "invalid option",
			name: "invalid",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			if testCase.opt != nil {
				AddOption(testCase.opt)
			}

			require.Equal(t, GetOption(testCase.name), testCase.opt)
		})
	}
}

func TestSet(t *testing.T) {
	cfg := &config{
		Name:  "cfg",
		value: "val",
	}
	Set(cfg.Name, cfg.value)

	require.Equal(t, cfg, Configs[cfg.Name])
}

func TestGet(t *testing.T) {
	type testCase struct {
		test      string
		name      string
		cfg       *config
		serverCtx *server
	}

	testCases := []testCase{
		{
			test: "valid cfg",
			name: "valid",
			cfg: &config{
				Name: "valid",
			},
		},
		{
			test: "invalid cfg",
			name: "invalid",
		},
		{
			test: "valid cfg in server ctx",
			name: "valid",
			cfg: &config{
				Name: "valid",
			},
			serverCtx: &server{},
		},
		{
			test:      "invalid cfg in server ctx",
			name:      "invalid",
			serverCtx: &server{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			if testCase.cfg != nil {
				Set(testCase.cfg.Name, testCase.cfg.value)
			}

			if testCase.serverCtx != nil {
				ServerContext = testCase.serverCtx
			}

			require.Equal(t, Get(testCase.name), testCase.cfg)
		})
	}
}

func TestHas(t *testing.T) {
	type testCase struct {
		test      string
		name      string
		cfg       *config
		serverCtx *server
		expectedResult  bool
	}

	testCases := []testCase{
		{
			test: "valid cfg",
			name: "valid",
			cfg: &config{
				Name: "valid",
			},
			expectedResult: true,
		},
		{
			test: "invalid cfg",
			name: "invalid",
			expectedResult: false,
		},
		{
			test: "valid cfg in server ctx",
			name: "valid",
			cfg: &config{
				Name: "valid",
			},
			serverCtx: &server{},
			expectedResult: true,
		},
		{
			test:      "invalid cfg in server ctx",
			name:      "invalid",
			serverCtx: &server{},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			if testCase.cfg != nil {
				Set(testCase.cfg.Name, testCase.cfg.value)
			}

			if testCase.serverCtx != nil {
				ServerContext = testCase.serverCtx
			}

			require.Equal(t, Has(testCase.name), testCase.expectedResult)
		})
	}
}
