package exec

import (
	"fmt"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Exec struct {
	// Configs contains all exec context vars used by Get and Set
	Configs map[string]*config
	// Tasks contains all exec tasks
	Tasks map[string]*task
	// Servers contains all exec servers
	Servers map[string]*server
	// TaskGroups contains all exec task groups
	TaskGroups map[string]*taskGroup
	// Arguments contains all exec arguments
	Arguments map[string]*Argument
	// Options contains all exec options
	Options map[string]*Option
	// ServerContext is the current active server
	ServerContext *server
	// TaskContext is the current executed task
	TaskContext *task

	before           map[string][]string
	after            map[string][]string
	serverContextF   func() []string //must return one server name
	argumentSequence int
}

// New returns a new *Exec instance
func New() *Exec {
	return &Exec{
		Configs:        make(map[string]*config),
		Tasks:          make(map[string]*task),
		Servers:        make(map[string]*server),
		TaskGroups:     make(map[string]*taskGroup),
		Arguments:      make(map[string]*Argument),
		Options:        make(map[string]*Option),
		before:         make(map[string][]string),
		after:          make(map[string][]string),
		serverContextF: func() []string { return nil },
	}
}

// Instance is the default empty exported instance of *Exec
// used to be able to create external recipes easily
var Instance = New()

// Init initializes the exec and executes the current command
// should be added to the end of all exec declarations
func (e *Exec) Init() {
	subtasks := make(map[string]*task)

	for name, task := range e.Tasks {
		task.Arguments = mergeArguments(task.removeArguments, e.Arguments, task.Arguments)
		task.Options = mergeOptions(task.removeOptions, e.Options, task.Options)

		if !task.private {
			subtasks[name] = task
		}
	}

	for name := range e.TaskGroups {
		e.TaskGroups[name].task.Arguments = mergeArguments(e.TaskGroups[name].task.removeArguments, e.Arguments, e.TaskGroups[name].task.Arguments)
		e.TaskGroups[name].task.Options = mergeOptions(e.TaskGroups[name].task.removeOptions, e.Options, e.TaskGroups[name].task.Options)
		e.Tasks[name] = e.TaskGroups[name].task
		subtasks[name] = e.TaskGroups[name].task
	}

	for _, task := range e.Tasks {
		if e.before[task.Name] != nil {
			for _, bt := range e.before[task.Name] {
				if e.Tasks[bt] != nil {
					task.before = append(task.before, e.Tasks[bt])
				}
			}
		}
		if e.after[task.Name] != nil {
			for _, at := range e.after[task.Name] {
				if e.Tasks[at] != nil {
					task.after = append(task.after, e.Tasks[at])
				}
			}
		}
	}

	var rootTask = task{
		subtasks: subtasks,
	}

	rootTask.Arguments = e.Arguments
	rootTask.Options = mergeOptions(map[string]string{}, e.Options, rootTask.Options)

	if err := run(&rootTask); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	} else {
		for _, s := range e.Servers {
			_ = s.sshClient.Close()
		}
	}
}

// NewArgument returns a new Argument
func (e *Exec) NewArgument(name string, description string) *Argument {
	var arg = &Argument{
		Name:        name,
		Description: description,
	}
	return arg
}

// AddArgument adds an Argument to exec
func (e *Exec) AddArgument(argument *Argument) {
	if _, ok := e.Arguments[argument.Name]; !ok {
		argument.sequence = e.argumentSequence
		e.argumentSequence++
		e.Arguments[argument.Name] = argument
	}
}

// GetArgument return an Argument pointer
func (e *Exec) GetArgument(name string) *Argument {
	if arg, ok := e.Arguments[name]; ok {
		return arg
	}

	return nil
}

// NewOption returns a new Option
func (e *Exec) NewOption(name string, description string) *Option {
	var opt = &Option{
		Name:        name,
		Description: description,
	}
	return opt
}

// AddOption adds an Option to exec
func (e *Exec) AddOption(option *Option) {
	if _, ok := e.Options[option.Name]; !ok {
		e.Options[option.Name] = option
	}
}

// GetOption return an Option pointer
func (e *Exec) GetOption(name string) *Option {
	if opt, ok := e.Options[name]; ok {
		return opt
	}

	return nil
}

// Set sets a exec Config
func (e *Exec) Set(name string, value interface{}) {
	e.Configs[name] = &config{Name: name, value: value}
}

// Get gets a Config value either set in a Server or directly in exec
func (e *Exec) Get(name string) *config {
	if e.ServerContext != nil {
		if c, ok := e.ServerContext.Configs[name]; ok {
			return c
		}
	}
	if c, ok := e.Configs[name]; ok {
		return c
	}
	return nil
}

// Has checks if a Config is available
func (e *Exec) Has(name string) bool {
	if e.ServerContext != nil {
		if _, ok := e.ServerContext.Configs[name]; ok {
			return true
		}
	}
	_, ok := e.Configs[name]
	return ok
}

// Server adds a new Server to exec
// dsn should be user@host:port
func (e *Exec) Server(name string, dsn string) *server {
	e.Servers[name] = &server{
		Name:      name,
		Dsn:       dsn,
		Configs:   make(map[string]*config),
		sshClient: &sshClient{},
	}
	return e.Servers[name]
}

// Task inherits the exec Arguments and can override and/or have new Options
// it accepts a name and a func; the func content is executed on each command execution
func (e *Exec) Task(name string, f func()) *task {
	e.Tasks[name] = &task{
		Name:            name,
		Arguments:       make(map[string]*Argument),
		Options:         make(map[string]*Option),
		exec:            e,
		removeArguments: make(map[string]string),
		removeOptions:   make(map[string]string),
		serverContextF: func() []string {
			return nil
		},
	}
	e.Tasks[name].run = func() {
		// set task context
		e.TaskContext = e.Tasks[name]

		run, onServers := e.shouldIRun()

		//skip tasks's server checking if requested
		if run && len(onServers) > 0 {
			for _, server := range e.Servers {
				for _, onServer := range onServers {
					if (server.Name == onServer || server.HasRole(onServer)) && e.Servers[onServer] != nil {
						// set server context
						e.ServerContext = server

						color.White("➤ Executing task %s on server %s", color.YellowString(name), color.GreenString(fmt.Sprintf("[%s]", server.Name)))

						//execute task's func
						f()

						//reset server context
						e.ServerContext = nil
					}
				}
			}

		} else if run && len(onServers) == 0 {
			color.White("➤ Executing task %s", color.YellowString(name))

			//execute task's func
			f()
		} else {
			e.taskNotAllowedToRunPrint(onServers, name)
		}

		//reset task context
		e.TaskContext = nil
	}
	return e.Tasks[name]
}

// TaskGroup inherits the exec Arguments and can override and/or have new Options
// and it will run all associated tasks
func (e *Exec) TaskGroup(name string, tasks ...string) *taskGroup {
	e.TaskGroups[name] = &taskGroup{
		Name: name,
		task: &task{
			Name:            name,
			removeArguments: make(map[string]string),
			removeOptions:   make(map[string]string),
			run: func() {
				color.White("➤ Executing task group %s", color.YellowString(name))
				for _, task := range tasks {
					if e.Tasks[task] == nil {
						continue
					}

					if e.Tasks[task].once && e.Tasks[task].executedOnce {
						continue
					}

					//set task context
					e.TaskContext = e.Tasks[task]

					e.Tasks[task].run()

					if e.Tasks[task].once && !e.Tasks[task].executedOnce {
						e.Tasks[task].executedOnce = true
					}

					//reset task context
					e.TaskContext = nil
				}
			},
		},
	}
	e.TaskGroups[name].tasks = append(e.TaskGroups[name].tasks, tasks...)
	return e.TaskGroups[name]
}

// Local runs a local command and displays/returns the output for further usage, for example in a Task func
func (e *Exec) Local(command string, args ...interface{}) (o Output) {
	command = e.Parse(fmt.Sprintf(command, args...))

	color.Green("[%s] %s %s", "local", ">", color.WhiteString("`%s`", command))

	cmd := exec.Command("/bin/sh", "-c", command)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		o.err = err
		color.Red("[%s] %s %q", "local", "<", o.err)
		return o
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		o.err = err
		color.Red("[%s] %s %q", "local", "<", o.err)
		return o
	}

	err = cmd.Start()
	if err != nil {
		o.err = err
		color.Red("[%s] %s %q", "local", "<", o.err)
		return o
	}

	output := ""
	buf := make([]byte, 1024)

	n, err := stdout.Read(buf)

	if err != nil {
		o.err = err
	} else {
		color.Green("[%s] %s\n", "local", "<")
		for _, v := range buf[:n] {
			fmt.Printf("%c", v)
		}
		output = string(buf[:n])
	}
	for err == nil {
		n, err = stdout.Read(buf)
		output += string(buf[:n])
		for _, v := range buf[:n] {
			fmt.Printf("%c", v)
		}
		if err != nil && err != io.EOF {
			o.err = err
		}
	}

	o.text = strings.TrimSpace(output)

	if len(o.text) == 0 {
		color.Red("[%s] %s\n", "local", "<")
		bytesB, _ := ioutil.ReadAll(stderr)
		fmt.Printf("%s\n", strings.TrimSpace(string(bytesB)))
	}

	err = cmd.Wait()
	if err != nil {
		color.Red("[%s] %s %q", "local", "<", err)
	}

	return o
}

// Println parses a text template, if founds a {{ var }}, it automatically runs the Get(var) on it
func (e *Exec) Println(text string) {
	fmt.Println(e.Parse(text))
}

// OnServers sets the server context dynamically
func (e *Exec) OnServers(f func() []string) {
	e.serverContextF = f
}

// Remote runs a command with args, in the ServerContext
func (e *Exec) Remote(command string, args ...interface{}) (o Output) {
	run, onServers := e.shouldIRun()

	if !run {
		e.commandNotAllowedToRunPrint(onServers, fmt.Sprintf(command, args...))
		return o
	}

	if e.ServerContext != nil {
		return e.remoteRun(fmt.Sprintf(command, args...), e.ServerContext)
	}

	return o
}

// Upload uploads a file or directory from local to remote, using native scp binary
func (e *Exec) Upload(local, remote string) {
	run, onServers := e.shouldIRun()

	if !run {
		e.commandNotAllowedToRunPrint(onServers, fmt.Sprintf("scp (local)%s > (remote)%s", local, remote))
	}

	var args = []string{"scp", "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r"}
	if e.ServerContext.key != nil {
		args = append(args, "-i "+*e.ServerContext.key)
	}
	args = append(args, local, e.ServerContext.Dsn+":"+remote)

	e.Local(strings.Join(args, " "))
}

// Download downloads a file or directory from remote to local, using native scp binary
func (e *Exec) Download(remote, local string) {
	run, onServers := e.shouldIRun()

	if !run {
		e.commandNotAllowedToRunPrint(onServers, fmt.Sprintf("scp (remote)%s > (local)%s", local, remote))
	}

	var args = []string{"scp", "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r"}
	if e.ServerContext.key != nil {
		args = append(args, "-i "+*e.ServerContext.key)
	}
	args = append(args, e.ServerContext.Dsn+":"+remote, local)

	e.Local(strings.Join(args, " "))
}

// Before sets tasks to run before task
func (e *Exec) Before(task string, tasksBefore ...string) {
	for _, tb := range tasksBefore {
		if !contains(e.before[task], tb) {
			e.before[task] = append(e.before[task], tb)
		}
	}
}

// After sets tasks to run after task
func (e *Exec) After(task string, tasksAfter ...string) {
	for _, ta := range tasksAfter {
		if !contains(e.after[task], ta) {
			e.after[task] = append(e.after[task], ta)
		}
	}
}

// remoteRun executes a command on a specific server
func (e *Exec) remoteRun(command string, server *server) (o Output) {
	e.ServerContext = server
	command = e.Parse(command)

	color.Green("[%s] %s %s", server.Name, ">", color.WhiteString("`%s`", command))

	if !server.sshClient.connOpened {
		err := server.sshClient.Connect(server.Dsn)
		if err != nil {
			color.Red("[%s] %s %q", "local", "<", err)
			o.err = err
		}
	}

	if server.sshClient.connOpened {
		err := server.sshClient.Run(command)
		if err != nil {
			o.err = err
			color.Red("[%s] %s %q", server.Name, "<", err)
		}

		output := ""
		buf := make([]byte, 1024)

		n, err := server.sshClient.remoteStdout.Read(buf)

		if err != nil {
			o.err = err
		} else {
			color.Green("[%s] %s\n", server.Name, "<")
			for _, v := range buf[:n] {
				fmt.Printf("%c", v)
			}
			output = string(buf[:n])
		}
		for err == nil {
			n, err = server.sshClient.remoteStdout.Read(buf)
			output += string(buf[:n])
			for _, v := range buf[:n] {
				fmt.Printf("%c", v)
			}
			if err != nil && err != io.EOF {
				o.err = err
			}
		}

		o.text = strings.TrimSpace(output)

		if len(o.text) == 0 {
			color.Red("[%s] %s\n", server.Name, "<")
			bytesB, _ := ioutil.ReadAll(server.sshClient.remoteStderr)
			fmt.Printf("%s\n", strings.TrimSpace(string(bytesB)))
		}

		err = server.sshClient.Wait()
		if err != nil {
			color.Red("[%s] %s %q", server.Name, "<", err)
		}
	}

	return o
}

func (e *Exec) shouldIRun() (run bool, onServers []string) {
	run = true

	//default values if serverContextF is set
	if s := e.serverContextF(); len(s) > 0 {
		onServers = s
	}

	//inside a task
	if e.TaskContext != nil {
		//task has a serverContextF
		if s := e.TaskContext.serverContextF(); len(s) > 0 {
			onServers = s
		}

		//task needs to run only on some servers
		if len(e.TaskContext.onlyOnServers) > 0 {
			run = false
			for _, oS := range onServers {
				for _, oOS := range e.TaskContext.onlyOnServers {
					//task on server matches only on servers
					if oS == oOS {
						run = true
					}
				}
			}
			onServers = e.TaskContext.onlyOnServers
		}

		if e.TaskContext.once && e.TaskContext.executedOnce {
			run = false
		}
	}

	return run, onServers
}

func (e *Exec) commandNotAllowedToRunPrint(onServers []string, command string) {
	fmt.Printf("%s%s%s\n", color.CyanString("[local] > Command `"), color.WhiteString(command), color.CyanString("` can run only on %s", onServers))
}

func (e *Exec) taskNotAllowedToRunPrint(onServers []string, task string) {
	fmt.Printf("%s%s%s\n", color.CyanString("[local] > Task `"), color.WhiteString(task), color.CyanString("` can run only on %s", onServers))
}

// onStart task setup
func (e *Exec) onStart() {
	if task, ok := e.Tasks["onStart"]; ok {
		task.run()
	}
}

// onEnd task setup
func (e *Exec) onEnd() {
	if task, ok := e.Tasks["onEnd"]; ok {
		task.run()
	}
}
