package exec

import (
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var (
	Configs          = make(map[string]*config)
	Tasks            = make(map[string]*task)
	Servers          = make(map[string]*server)
	TaskGroups       = make(map[string]*taskGroup)
	Arguments        = make(map[string]*Argument)
	Options          = make(map[string]*Option)
	ServerContext    *server
	serverContextF   = func() string { return "" } //must return one server name
	TaskContext      *task
	argumentSequence int = 0
)

// Init initializes the exec and executes the current command
// should be added to the end of all exec declarations
func Init() {
	subtasks := make(map[string]*task)

	for name, task := range Tasks {
		task.Arguments = Arguments
		task.Options = mergeOptions(Options, task.Options)

		if !task.private {
			subtasks[name] = task
		}
	}

	for name, taskGroup := range TaskGroups {
		taskGroup.task.run = func() {
			for _, task := range taskGroup.tasks {
				if Tasks[task] == nil {
					continue
				}

				if Tasks[task].once && Tasks[task].executedOnce {
					continue
				}

				TaskContext = Tasks[task]
				Tasks[task].run()

				if Tasks[task].once && !Tasks[task].executedOnce {
					Tasks[task].executedOnce = true
				}
			}
		}
		taskGroup.task.Arguments = Arguments
		taskGroup.task.Options = mergeOptions(Options, taskGroup.task.Options)
		Tasks[name] = taskGroup.task
		subtasks[name] = taskGroup.task
	}

	var rootTask = task{
		subtasks: subtasks,
	}
	rootTask.Arguments = Arguments
	rootTask.Options = mergeOptions(Options, rootTask.Options)

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
		argumentSequence += 1
		Arguments[argument.Name] = argument
	}
}

func GetArgument(name string) *Argument {
	if arg, ok := Arguments[name]; ok {
		return arg
	} else {
		return nil
	}
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
func Server(name string, host string) *server {
	Servers[name] = &server{
		Name:      name,
		Host:      host,
		Configs:   make(map[string]*config),
		sshClient: &sshClient{},
	}
	return Servers[name]
}

// Task inherits the exec Arguments and can override and/or have new Options
// it accepts a name and a func; the func content is executed on each command execution
func Task(name string, f func()) *task {
	Tasks[name] = &task{
		Name:           name,
		Arguments:      make(map[string]*Argument),
		Options:        make(map[string]*Option),
		serverContextF: func() string { return "" },
		run: func() {
			color.White("➤ Executing task %s", color.YellowString(name))
			f()
		}}
	return Tasks[name]
}

// TaskGroup inherits the exec Arguments and can override and/or have new Options
// and it will run all associated tasks
func TaskGroup(name string, tasks ...string) *taskGroup {
	TaskGroups[name] = &taskGroup{
		Name: name,
		task: &task{},
	}
	TaskGroups[name].tasks = append(TaskGroups[name].tasks, tasks...)
	return TaskGroups[name]
}

// Local runs a local command and displays/returns the output for further usage, for example in a Task func
func Local(command string) (o output) {
	command = Parse(command)

	color.Green("[%s] %s %s", "local", ">", color.WhiteString(command))

	cs := strings.SplitN(command, " ", 2)

	output, err := exec.Command(cs[0], strings.Split(cs[1], " ")...).CombinedOutput()
	o.text = strings.TrimSpace(string(output))
	if err != nil {
		color.Red("[%s] %s %s", "local", "<", o.text)
	} else {
		color.Green("[%s] %s\n%s", "local", "<", color.WhiteString(o.text))
	}
	return o
}

// Println parses a text template, if founds a {{ var }}, it automatically runs the Get(var) on it
func Println(text string) {
	fmt.Println(Parse(text))
}

// Cd is a remote helper function that runs a `cd` before a command
func Cd(path string) {
	command := "cd " + Parse(path)
	color.Green("[%s] %s %s", ServerContext.Name, color.GreenString(">"), command)
	ServerContext.sshClient.env = command + "; "
}

// CommandExists is a remote helper function that checks if a command exists on server
func CommandExist(command string) bool {
	return Remote(fmt.Sprintf("if hash %s 2>/dev/null; then echo 'true'; fi", command)).Bool()
}

// OnServer sets the server context dynamically
func OnServer(f func() string) {
	serverContextF = f
}

// RemoteRun executes a command on a specific server
func RemoteRun(command string, server *server) (o output) {
	ServerContext = server
	command = Parse(command)

	color.Green("[%s] %s %s", server.Name, ">", color.WhiteString(command))

	if !server.sshClient.connOpened {
		err := server.sshClient.Connect(server.Host)
		if err != nil {
			color.Red("[%s] %s %s", "local", "<", err)
			o.err = err
		}
	}

	if server.sshClient.connOpened {
		err := server.sshClient.Run(command)
		if err != nil {
			color.Red("[%s] %s %s", server.Name, "<", err)
		}
		output, err := ioutil.ReadAll(server.sshClient.Stdout())

		execErr := server.sshClient.Wait()

		if execErr != nil {
			o.err = execErr
		}

		o.text = strings.TrimSpace(string(output))
		if o.text != "" && o.err == nil {
			color.Green("[%s] %s\n%s", server.Name, "<", color.WhiteString(o.text))
		} else if o.text != "" && o.err != nil {
			color.Red("[%s] %s %s", server.Name, "<", o.String())
		}
	}

	return o
}

// Remote runs a command on one server
func Remote(command string) (o output) {
	run, onServer := shouldIRun()

	if run && Servers[onServer] != nil {
		return RemoteRun(command, Servers[onServer])
	} else {
		notAllowedForPrint(onServer, command)
		return o
	}
}

// Remotes runs a command on servers with a specific role or name key
func Remotes(command string) {
	run, onServer := shouldIRun()

	for name, server := range Servers {
		if server.HasRole(onServer) || name == onServer {
			if run {
				go RemoteRun(command, server)
			} else {
				notAllowedForPrint(onServer, command)
			}
		}
	}
}

// Upload uploads a file from local to remote, using native scp binary
func Upload(local, remote string) {
	run, onServer := shouldIRun()

	if run && Servers[onServer] != nil {
		var args = []string{"scp"}
		if Servers[onServer].key != nil {
			args = append(args, "-i "+*Servers[onServer].key)
		}
		args = append(args, local, Servers[onServer].Host+":"+remote)

		Local(strings.Join(args, " "))
	} else {
		fmt.Printf("%s%s%s\n", color.CyanString("[%s] %s Uploading `", getOnServerForPrint(onServer), ">"), color.WhiteString(local), color.CyanString("` not allowed to run"))
		color.Cyan("Reasons: onServer AND/OR onlyOnServers is not met")
	}
}

// Download downloads a file from remote to local, using native scp binary
func Download(remote, local string) {
	run, onServer := shouldIRun()

	if run && Servers[onServer] != nil {
		var args = []string{"scp"}
		if Servers[onServer].key != nil {
			args = append(args, "-i "+*Servers[onServer].key)
		}
		args = append(args, Servers[onServer].Host+":"+remote, local)

		Local(strings.Join(args, " "))
	} else {
		fmt.Printf("%s%s%s\n", color.CyanString("[%s] %s Uploading `", getOnServerForPrint(onServer), ">"), color.WhiteString(local), color.CyanString("` not allowed to run"))
		color.Cyan("Reasons: onServer AND/OR onlyOnServers is not met")
	}
}

func shouldIRun() (run bool, onServer string) {
	//default values if server context is set
	if s := serverContextF(); s != "" {
		run = true
		onServer = s
	}

	//inside a task
	if TaskContext != nil {
		//task has a server
		if s := TaskContext.serverContextF(); s != "" {
			run = true
			onServer = s
		}

		//task needs to run only on some servers
		if len(TaskContext.onlyOnServers) > 0 {
			run = false
			for _, sN := range TaskContext.onlyOnServers {
				//task on server matches only on servers
				if onServer == sN {
					run = true
				}
			}
		}
	}

	return run, onServer
}

func getOnServerForPrint(onServer string) string {
	if onServer != "" {
		return onServer
	} else {
		return "?"
	}
}

func notAllowedForPrint(onServer, command string) {
	fmt.Printf("%s%s%s\n", color.CyanString("[%s] %s Command `", getOnServerForPrint(onServer), ">"), color.WhiteString(command), color.CyanString("` not allowed to run"))
	color.Cyan("Reasons: onServer AND/OR onlyOnServers is not met")
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
