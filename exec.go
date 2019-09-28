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

var (
	// Configs contains all exec context vars used by Get and Set
	Configs = make(map[string]*config)
	// Tasks contains all exec tasks
	Tasks = make(map[string]*task)
	// Servers contains all exec servers
	Servers = make(map[string]*server)
	// TaskGroups contains all exec task groups
	TaskGroups = make(map[string]*taskGroup)
	// Arguments contains all exec arguments
	Arguments = make(map[string]*Argument)
	// Options contains all exec options
	Options = make(map[string]*Option)
	// ServerContext is the current active server
	ServerContext *server
	// TaskContext is the current executed task
	TaskContext *task

	before           = make(map[string][]string)
	after            = make(map[string][]string)
	serverContextF   = func() []string { return nil } //must return one server name
	argumentSequence int
)

// Init initializes the exec and executes the current command
// should be added to the end of all exec declarations
func Init() {
	subtasks := make(map[string]*task)

	for name, task := range Tasks {
		task.Arguments = mergeArguments(task.removeArguments, Arguments, task.Arguments)
		task.Options = mergeOptions(task.removeOptions, Options, task.Options)

		if !task.private {
			subtasks[name] = task
		}
	}

	for name := range TaskGroups {
		TaskGroups[name].task.Arguments = mergeArguments(TaskGroups[name].task.removeArguments, Arguments, TaskGroups[name].task.Arguments)
		TaskGroups[name].task.Options = mergeOptions(TaskGroups[name].task.removeOptions, Options, TaskGroups[name].task.Options)
		Tasks[name] = TaskGroups[name].task
		subtasks[name] = TaskGroups[name].task
	}

	for _, task := range Tasks {
		if before[task.Name] != nil {
			for _, bt := range before[task.Name] {
				if Tasks[bt] != nil {
					task.before = append(task.before, Tasks[bt])
				}
			}
		}
		if after[task.Name] != nil {
			for _, at := range after[task.Name] {
				if Tasks[at] != nil {
					task.after = append(task.after, Tasks[at])
				}
			}
		}
	}

	var rootTask = task{
		subtasks: subtasks,
	}

	rootTask.Arguments = Arguments
	rootTask.Options = mergeOptions(map[string]string{}, Options, rootTask.Options)

	if err := run(&rootTask); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		for _, s := range Servers {
			s.sshClient.Close()
		}
	}
}

// NewArgument returns a new Argument
func NewArgument(name string, description string) *Argument {
	var arg = &Argument{
		Name:        name,
		Description: description,
	}
	return arg
}

// AddArgument adds an Argument to exec
func AddArgument(argument *Argument) {
	if _, ok := Arguments[argument.Name]; !ok {
		argument.sequence = argumentSequence
		argumentSequence++
		Arguments[argument.Name] = argument
	}
}

// GetArgument return an Argument pointer
func GetArgument(name string) *Argument {
	if arg, ok := Arguments[name]; ok {
		return arg
	}

	return nil
}

// NewOption returns a new Option
func NewOption(name string, description string) *Option {
	var opt = &Option{
		Name:        name,
		Description: description,
	}
	return opt
}

// AddOption adds an Option to exec
func AddOption(option *Option) {
	if _, ok := Options[option.Name]; !ok {
		Options[option.Name] = option
	}
}

// GetOption return an Option pointer
func GetOption(name string) *Option {
	if opt, ok := Options[name]; ok {
		return opt
	}

	return nil
}

// Set sets a exec Config
func Set(name string, value interface{}) {
	Configs[name] = &config{Name: name, value: value}
}

// Get gets a Config value either set in a Server or directly in exec
func Get(name string) *config {
	if ServerContext != nil {
		if c, ok := ServerContext.Configs[name]; ok {
			return c
		}
	}
	if c, ok := Configs[name]; ok {
		return c
	}
	return nil
}

// Has checks if a Config is available
func Has(name string) bool {
	if ServerContext != nil {
		if _, ok := ServerContext.Configs[name]; ok {
			return true
		}
	}
	_, ok := Configs[name]
	return ok
}

// Server adds a new Server to exec
// dsn should be user@host:port
func Server(name string, dsn string) *server {
	Servers[name] = &server{
		Name:      name,
		Dsn:       dsn,
		Configs:   make(map[string]*config),
		sshClient: &sshClient{},
	}
	return Servers[name]
}

// Task inherits the exec Arguments and can override and/or have new Options
// it accepts a name and a func; the func content is executed on each command execution
func Task(name string, f func()) *task {
	Tasks[name] = &task{
		Name:            name,
		Arguments:       make(map[string]*Argument),
		Options:         make(map[string]*Option),
		removeArguments: make(map[string]string),
		removeOptions:   make(map[string]string),
		serverContextF: func() []string {
			return nil
		},
	}
	Tasks[name].run = func() {
		// set task context
		TaskContext = Tasks[name]

		run, onServers := shouldIRun()

		//skip tasks's server checking if requested
		if run && len(onServers) > 0 {
			for _, server := range Servers {
				for _, onServer := range onServers {
					if (server.Name == onServer || server.HasRole(onServer)) && Servers[onServer] != nil {
						// set server context
						ServerContext = server

						color.White("➤ Executing task %s on server %s", color.YellowString(name), color.GreenString(fmt.Sprintf("[%s]", server.Name)))

						//execute task's func
						f()

						//reset server context
						ServerContext = nil
					}
				}
			}

		} else if run && len(onServers) == 0 {
			color.White("➤ Executing task %s", color.YellowString(name))

			//execute task's func
			f()
		} else {
			taskNotAllowedToRunPrint(onServers, name)
		}

		//reset task context
		TaskContext = nil
	}
	return Tasks[name]
}

// TaskGroup inherits the exec Arguments and can override and/or have new Options
// and it will run all associated tasks
func TaskGroup(name string, tasks ...string) *taskGroup {
	TaskGroups[name] = &taskGroup{
		Name: name,
		task: &task{
			Name:            name,
			removeArguments: make(map[string]string),
			removeOptions:   make(map[string]string),
			run: func() {
				color.White("➤ Executing task group %s", color.YellowString(name))
				for _, task := range tasks {
					if Tasks[task] == nil {
						continue
					}

					if Tasks[task].once && Tasks[task].executedOnce {
						continue
					}

					//set task context
					TaskContext = Tasks[task]

					Tasks[task].run()

					if Tasks[task].once && !Tasks[task].executedOnce {
						Tasks[task].executedOnce = true
					}

					//reset task context
					TaskContext = nil
				}
			},
		},
	}
	TaskGroups[name].tasks = append(TaskGroups[name].tasks, tasks...)
	return TaskGroups[name]
}

// Local runs a local command and displays/returns the output for further usage, for example in a Task func
func Local(command string, args ...interface{}) (o output) {
	command = Parse(fmt.Sprintf(command, args...))

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

	o.text = strings.TrimSpace(string(output))

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
func Println(text string) {
	fmt.Println(Parse(text))
}

// OnServers sets the server context dynamically
func OnServers(f func() []string) {
	serverContextF = f
}

// RemoteRun executes a command on a specific server
func RemoteRun(command string, server *server) (o output) {
	ServerContext = server
	command = Parse(command)

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

// Remote runs a command with args, in the ServerContext
func Remote(command string, args ...interface{}) (o output) {
	run, onServers := shouldIRun()

	if !run {
		commandNotAllowedToRunPrint(onServers, fmt.Sprintf(command, args...))
		return o
	}

	if ServerContext != nil {
		return RemoteRun(fmt.Sprintf(command, args...), ServerContext)
	}

	return o
}

// Upload uploads a file or directory from local to remote, using native scp binary
func Upload(local, remote string) {
	run, onServers := shouldIRun()

	if !run {
		commandNotAllowedToRunPrint(onServers, fmt.Sprintf("scp (local)%s > (remote)%s", local, remote))
	}

	var args = []string{"scp", "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r"}
	if ServerContext.key != nil {
		args = append(args, "-i "+*ServerContext.key)
	}
	args = append(args, local, ServerContext.Dsn+":"+remote)

	Local(strings.Join(args, " "))
}

// Download downloads a file or directory from remote to local, using native scp binary
func Download(remote, local string) {
	run, onServers := shouldIRun()

	if !run {
		commandNotAllowedToRunPrint(onServers, fmt.Sprintf("scp (remote)%s > (local)%s", local, remote))
	}

	var args = []string{"scp", "-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -r"}
	if ServerContext.key != nil {
		args = append(args, "-i "+*ServerContext.key)
	}
	args = append(args, ServerContext.Dsn+":"+remote, local)

	Local(strings.Join(args, " "))
}

// Before sets tasks to run before task
func Before(task string, tasksBefore ...string) {
	for _, tb := range tasksBefore {
		if !contains(before[task], tb) {
			before[task] = append(before[task], tb)
		}
	}
}

// After sets tasks to run after task
func After(task string, tasksAfter ...string) {
	for _, ta := range tasksAfter {
		if !contains(after[task], ta) {
			after[task] = append(after[task], ta)
		}
	}
}

func shouldIRun() (run bool, onServers []string) {
	run = true

	//default values if serverContextF is set
	if s := serverContextF(); len(s) > 0 {
		onServers = s
	}

	//inside a task
	if TaskContext != nil {
		//task has a serverContextF
		if s := TaskContext.serverContextF(); len(s) > 0 {
			onServers = s
		}

		//task needs to run only on some servers
		if len(TaskContext.onlyOnServers) > 0 {
			run = false
			for _, oS := range onServers {
				for _, oOS := range TaskContext.onlyOnServers {
					//task on server matches only on servers
					if oS == oOS {
						run = true
					}
				}
			}
			onServers = TaskContext.onlyOnServers
		}

		if TaskContext.once && TaskContext.executedOnce {
			run = false
		}
	}

	return run, onServers
}

func commandNotAllowedToRunPrint(onServers []string, command string) {
	fmt.Printf("%s%s%s\n", color.CyanString("[local] > Command `"), color.WhiteString(command), color.CyanString("` can run only on %s", onServers))
}

func taskNotAllowedToRunPrint(onServers []string, task string) {
	fmt.Printf("%s%s%s\n", color.CyanString("[local] > Task `"), color.WhiteString(task), color.CyanString("` can run only on %s", onServers))
}

// onStart task setup
func onStart() {
	if task, ok := Tasks["onStart"]; ok {
		task.run()
	}
}

// onEnd task setup
func onEnd() {
	if task, ok := Tasks["onEnd"]; ok {
		task.run()
	}
}
