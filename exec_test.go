package exec

import (
	"github.com/go-exec/exec/ssh_mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	e := New()

	require.IsType(t, &Exec{}, e)
}

func TestExec_NewArgument(t *testing.T) {
	e := New()

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
	e := New()

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
			e := New()

			if testCase.arg != nil {
				e.AddArgument(testCase.arg)
			}

			require.Equal(t, e.GetArgument(testCase.name), testCase.arg)
		})
	}
}

func TestExec_NewOption(t *testing.T) {
	e := New()

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
	e := New()

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
			e := New()

			if testCase.opt != nil {
				e.AddOption(testCase.opt)
			}

			require.Equal(t, e.GetOption(testCase.name), testCase.opt)
		})
	}
}

func TestExec_Set(t *testing.T) {
	e := New()

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
			e := New()

			if testCase.cfg != nil {
				e.Set(testCase.cfg.Name, testCase.cfg.value)
			}

			if testCase.serverCtx != nil {
				e.ServerContext = testCase.serverCtx
			}

			require.Equal(t, e.Get(testCase.name), testCase.cfg)
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
			e := New()

			if testCase.cfg != nil {
				e.Set(testCase.cfg.Name, testCase.cfg.value)
			}

			if testCase.serverCtx != nil {
				e.ServerContext = testCase.serverCtx
			}

			require.Equal(t, e.Has(testCase.name), testCase.expectedResult)
		})
	}
}

func TestExec_Server(t *testing.T) {
	e := New()

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
	e := New()

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
	e := New()

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
			e := New()

			e.Before(testCase.task.Name, testCase.before...)

			require.Contains(t, e.before, testCase.task.Name)
			require.Equal(t, len(e.before[testCase.task.Name]), testCase.unique)
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
			e := New()

			e.Before(testCase.task.Name, testCase.after...)

			require.Contains(t, e.before, testCase.task.Name)
			require.Equal(t, len(e.before[testCase.task.Name]), testCase.unique)
		})
	}
}

func TestExec_Remote(t *testing.T) {
	type args struct {
		command string
		args    []interface{}
	}
	tests := []struct {
		name  string
		args  args
		wantO Output
	}{
		{
			name: "test",
			args: args{
				command: `echo hello`,
			},
			wantO: Output{
				text: "hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := ssh_mock.NewServer(t)
			defer server.Shutdown()
			conn := server.Dial(ssh_mock.ClientConfig())
			defer conn.Close()

			e := New()

			s := e.Server("mock", "")

			s.sshClient.WithConnection(conn)

			e.ServerContext = s

			gotO := e.Remote(tt.args.command, tt.args.args...)

			require.Equal(t, tt.wantO, gotO, "Remote() = %v, want %v", gotO, tt.wantO)
		})
	}
}
