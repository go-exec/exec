package exec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func setupTestCase(t *testing.T) (*Exec, func(t *testing.T)) {
	return New(), func(t *testing.T) {}
}

func TestNew(t *testing.T) {
	e, teardown := setupTestCase(t)
	defer teardown(t)

	require.IsType(t, &Exec{}, e)
}

func TestExec_NewArgument(t *testing.T) {
	e, teardown := setupTestCase(t)
	defer teardown(t)

	arg := &Argument{
		Name:        "name",
		Type:        0,
		Default:     nil,
		Multiple:    false,
		Description: "description",
		Value:       nil,
	}

	require.Equal(t, e.NewArgument(arg.Name, arg.Description), arg)
}

func TestExec_AddArgument(t *testing.T) {
	e, teardown := setupTestCase(t)
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

func TestExec_GetArgument(t *testing.T) {
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
			e, teardown := setupTestCase(t)

			if testCase.arg != nil {
				e.AddArgument(testCase.arg)
			}

			require.Equal(t, e.GetArgument(testCase.name), testCase.arg)

			teardown(t)
		})
	}
}

func TestExec_NewOption(t *testing.T) {
	e, teardown := setupTestCase(t)
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

func TestExec_AddOption(t *testing.T) {
	e, teardown := setupTestCase(t)
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

func TestExec_GetOption(t *testing.T) {
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
			e, teardown := setupTestCase(t)

			if testCase.opt != nil {
				e.AddOption(testCase.opt)
			}

			require.Equal(t, e.GetOption(testCase.name), testCase.opt)

			teardown(t)
		})
	}
}

func TestExec_Set(t *testing.T) {
	e, teardown := setupTestCase(t)
	defer teardown(t)

	cfg := &config{
		Name:  "cfg",
		value: "val",
	}
	e.Set(cfg.Name, cfg.value)

	require.Equal(t, cfg, e.Configs[cfg.Name])
}

func TestExec_Get(t *testing.T) {
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
			e, teardown := setupTestCase(t)

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

func TestExec_Has(t *testing.T) {
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
			e, teardown := setupTestCase(t)

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

func TestExec_Server(t *testing.T) {
	e, teardown := setupTestCase(t)
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

func TestExec_Task(t *testing.T) {
	e, teardown := setupTestCase(t)
	defer teardown(t)

	task := &task{
		Name:            "task",
		Arguments:       make(map[string]*Argument),
		Options:         make(map[string]*Option),
		exec:            e,
		removeArguments: make(map[string]string),
		removeOptions:   make(map[string]string),
	}
	e.Task(task.Name, func() {})

	require.Contains(t, e.Tasks, task.Name)
	require.Equal(t, task.exec, e.Tasks[task.Name].exec)
}

func TestExec_TaskGroup(t *testing.T) {
	e, teardown := setupTestCase(t)
	defer teardown(t)

	taskGroup := &taskGroup{
		Name: "taskGroup",
	}
	e.TaskGroup(taskGroup.Name)

	require.Contains(t, e.TaskGroups, taskGroup.Name)
	require.Equal(t, e.TaskGroups[taskGroup.Name].task.exec, e)
}

func TestExec_Before(t *testing.T) {
	type testCase struct {
		test   string
		task   *task
		before []string
		unique int
	}

	testCases := []testCase{
		{
			test: "valid",
			task: &task{
				Name: "task",
			},
			before: []string{"before 1"},
			unique: 1,
		},
		{
			test: "valid with unique before items",
			task: &task{
				Name: "task",
			},
			before: []string{"before 1", "before 2", "before 1"},
			unique: 2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			e, teardown := setupTestCase(t)

			e.Before(testCase.task.Name, testCase.before...)

			require.Contains(t, e.before, testCase.task.Name)
			require.Equal(t, len(e.before[testCase.task.Name]), testCase.unique)

			defer teardown(t)
		})
	}
}

func TestExec_After(t *testing.T) {
	type testCase struct {
		test   string
		task   *task
		after  []string
		unique int
	}

	testCases := []testCase{
		{
			test: "valid",
			task: &task{
				Name: "task",
			},
			after:  []string{"after 1"},
			unique: 1,
		},
		{
			test: "valid with unique after items",
			task: &task{
				Name: "task",
			},
			after:  []string{"after 1", "after 2", "after 1"},
			unique: 2,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.test, func(t *testing.T) {
			e, teardown := setupTestCase(t)

			e.Before(testCase.task.Name, testCase.after...)

			require.Contains(t, e.before, testCase.task.Name)
			require.Equal(t, len(e.before[testCase.task.Name]), testCase.unique)

			defer teardown(t)
		})
	}
}
