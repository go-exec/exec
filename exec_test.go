package exec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var e *Exec

func setupTestCase(t *testing.T) func(t *testing.T) {
	e = New()
	return func(t *testing.T) {
		t.Log("teardown test case")
	}
}

func TestNewArgument(t *testing.T) {
	teardown := setupTestCase(t)
	defer teardown(t)

	arg := &Argument{
		Name:        "name",
		Type:        0,
		Default:     nil,
		Multiple:    false,
		Description: "description",
		Value:       nil,
	}

	require.Equal(t, e.NewArgument(arg.Name, arg.Description), arg, "err")
}

func TestAddArgument(t *testing.T) {
	teardown := setupTestCase(t)
	defer teardown(t)

	arg := &Argument{
		Name:        "test",
		Type:        0,
		Default:     nil,
		Multiple:    false,
		Description: "",
		Value:       nil,
	}
	e.AddArgument(arg)

	require.Equal(t, arg, e.Arguments[arg.Name])
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
			teardown := setupTestCase(t)

			if testCase.arg != nil {
				e.AddArgument(testCase.arg)
			}

			require.Equal(t, e.GetArgument(testCase.name), testCase.arg)

			teardown(t)
		})
	}
}

func TestNewOption(t *testing.T) {
	teardown := setupTestCase(t)
	defer teardown(t)

	opt := &Option{
		Name:        "name",
		Type:        0,
		Default:     nil,
		Description: "description",
		Value:       nil,
	}
	require.Equal(t, e.NewOption(opt.Name, opt.Description), opt)
}

func TestAddOption(t *testing.T) {
	teardown := setupTestCase(t)
	defer teardown(t)

	opt := &Option{
		Name:        "name",
		Type:        0,
		Default:     nil,
		Description: "description",
		Value:       nil,
	}
	e.AddOption(opt)

	require.Equal(t, opt, e.Options[opt.Name])
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
			teardown := setupTestCase(t)

			if testCase.opt != nil {
				e.AddOption(testCase.opt)
			}

			require.Equal(t, e.GetOption(testCase.name), testCase.opt)

			teardown(t)
		})
	}
}

func TestSet(t *testing.T) {
	teardown := setupTestCase(t)
	defer teardown(t)

	cfg := &config{
		Name:  "cfg",
		value: "val",
	}
	e.Set(cfg.Name, cfg.value)

	require.Equal(t, cfg, e.Configs[cfg.Name])
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
			teardown := setupTestCase(t)

			if testCase.cfg != nil {
				e.Set(testCase.cfg.Name, testCase.cfg.value)
			}

			if testCase.serverCtx != nil {
				e.ServerContext = testCase.serverCtx
			}

			require.Equal(t, e.Get(testCase.name), testCase.cfg)

			teardown(t)
		})
	}
}

func TestHas(t *testing.T) {
	type testCase struct {
		test           string
		name           string
		cfg            *config
		serverCtx      *server
		expectedResult bool
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
			test:           "invalid cfg",
			name:           "invalid",
			expectedResult: false,
		},
		{
			test: "valid cfg in server ctx",
			name: "valid",
			cfg: &config{
				Name: "valid",
			},
			serverCtx:      &server{},
			expectedResult: true,
		},
		{
			test:           "invalid cfg in server ctx",
			name:           "invalid",
			serverCtx:      &server{},
			expectedResult: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			teardown := setupTestCase(t)

			if testCase.cfg != nil {
				e.Set(testCase.cfg.Name, testCase.cfg.value)
			}

			if testCase.serverCtx != nil {
				e.ServerContext = testCase.serverCtx
			}

			require.Equal(t, e.Has(testCase.name), testCase.expectedResult)

			teardown(t)
		})
	}
}

func TestServer(t *testing.T) {
	teardown := setupTestCase(t)
	defer teardown(t)

	cfg := &server{
		Name:      "server",
		Dsn:       "user@host:port",
		Configs:   make(map[string]*config),
		sshClient: &sshClient{},
	}
	e.Server(cfg.Name, cfg.Dsn)

	require.Equal(t, cfg, e.Servers[cfg.Name])
}
