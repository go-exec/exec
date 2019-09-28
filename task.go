package exec

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"sort"
	"strconv"
	"strings"
)

// errInvalidParams can be returned by a taskFunction to automatically
// display help text.
var errInvalidParams = errors.New("invalid parameters")

type task struct {
	Name      string
	Options   map[string]*Option
	Arguments map[string]*Argument

	run              taskFunction
	subtasks         map[string]*task
	shortDescription string
	description      string
	once             bool
	executedOnce     bool
	private          bool
	onlyOnServers    []string
	serverContextF   func() []string
	before           []*task
	after            []*task
	removeArguments  map[string]string
	removeOptions    map[string]string
}

type taskFunction func()

func (t *task) ShortDescription(description string) *task {
	t.shortDescription = description
	return t
}

func (t *task) Description(description string) *task {
	t.description = description
	return t
}

func (t *task) AddOption(option *Option) *task {
	t.Options[option.Name] = option
	return t
}

func (t *task) RemoveOption(optName string) *task {
	t.removeOptions[optName] = optName
	return t
}

func (t *task) AddArgument(arg *Argument) *task {
	t.Arguments[arg.Name] = arg
	return t
}

func (t *task) RemoveArgument(argName string) *task {
	t.removeArguments[argName] = argName
	return t
}

func (t *task) HasArgument(name string) bool {
	_, ok := t.Arguments[name]
	return ok
}

func (t *task) HasOption(name string) bool {
	_, ok := t.Options[name]
	return ok
}

func (t *task) GetArgument(name string) (arg *Argument) {
	if t.HasArgument(name) {
		return t.Arguments[name]
	}
	return arg
}

func (t *task) GetOption(name string) (opt *Option) {
	if t.HasOption(name) {
		return t.Options[name]
	}
	return opt
}

func (t *task) Once() *task {
	t.once = true
	return t
}

func (t *task) Private() *task {
	t.private = true
	return t
}

func (t *task) OnlyOnServers(servers []string) *task {
	t.onlyOnServers = servers
	return t
}

func (t *task) OnServers(f func() []string) *task {
	t.serverContextF = f
	return t
}

func (t *task) getOrderedArguments() sortArguments {
	var args sortArguments
	for _, argument := range t.Arguments {
		args = append(args, argument)
	}
	sort.Sort(args)
	return args
}

// printhelp prints the return value of help to the standard output.
func (t *task) printhelp(taskName string) {
	fmt.Printf(t.help(taskName))
}

// usageString returns a short string containing the syntax of this command.
// Command name should be set to one of the return values of FindCommand.
func (t *task) usageString(taskName string) string {
	rw := color.YellowString("Usage:\n    ") + taskName
	if len(t.subtasks) > 0 {
		rw += " <subcommand>"
	}
	rw += " [<options>]"
	for _, arg := range t.getOrderedArguments() {
		rw += fmt.Sprintf(" <%s>", arg.Name)
	}
	return rw
}

// help returns the full help text for this command  The text contains the
// syntax of the command, a description, a list of accepted options and
// arguments and available subcommands. Command name should be set to one of
// the return values of FindCommand.
func (t *task) help(taskName string) string {
	var rv string

	usage := t.usageString(taskName)
	rv += fmt.Sprintf("%s - %s\n", usage, t.shortDescription)

	if len(t.Options) > 0 {
		rv += color.YellowString("\nOptions:\n")
		for _, opt := range t.Options {
			rv += fmt.Sprintf("    %-20s %s\n", color.GreenString(opt.Explain()), opt.Description)
		}
	}

	if len(t.Arguments) > 0 {
		rv += color.YellowString("\nArguments:\n")
		for _, arg := range t.getOrderedArguments() {
			rv += fmt.Sprintf("    %-20s %s\n", color.GreenString(arg.Explain()), arg.Description)
		}
	}

	if len(t.subtasks) > 0 {
		rv += color.YellowString("\nAvailable commands:\n")
		var subtaskNames []string
		for subtaskName := range t.subtasks {
			subtaskNames = append(subtaskNames, subtaskName)
		}
		sort.Strings(subtaskNames)
		for _, subtaskName := range subtaskNames {
			rv += fmt.Sprintf("    %-20s %s\n", color.GreenString(subtaskName), t.subtasks[subtaskName].shortDescription)
		}
	}

	if len(t.description) > 0 {
		rv += fmt.Sprintln("\nDescription:")
		desc := strings.Trim(t.description, "\n")
		for _, line := range strings.Split(desc, "\n") {
			rv += fmt.Sprintf("    %s\n", line)
		}
	}

	return rv
}

// Execute runs the command. Command name is used to generate the help texts
// and should usually be set to one of the return values of FindCommand. The
// array of the arguments provided for this subcommand is used to generate the
// context and should be set to one of the return values of FindCommand as
// well. The command will not be executed with an insufficient number of
// arguments so there is no need to check that in the run function.
func (t *task) execute(taskName string, cmdArgs []string) error {
	err := t.parseArgs(cmdArgs)
	if err != nil {
		t.printhelp(t.Name)
		return err
	}

	// Is there a help flag and is it set?
	if help, ok := t.Options["help"]; ok && help.Bool() {
		t.printhelp(taskName)
		return nil
	}

	// Is this command only used to hold subcommands?
	if t.run == nil {
		t.printhelp(t.Name)
		return nil
	}

	// Execute it only once if requested
	if t.once && t.executedOnce {
		return nil
	}

	// Executing the onStart task
	onStart()

	if len(t.before) > 0 {
		for _, tb := range t.before {
			tb.run()
		}
	}

	// Runs the task's func
	t.run()

	if len(t.after) > 0 {
		for _, ta := range t.after {
			ta.run()
		}
	}

	// Executing the onEnd task
	onEnd()

	// Execute it only once if requested
	if t.once && !t.executedOnce {
		t.executedOnce = true
	}

	return nil
}

func (t *task) parseArgs(args []string) error {
	flagset := flag.NewFlagSet("sth", flag.ContinueOnError)
	flagset.Usage = func() {}

	for _, option := range t.Options {
		switch option.Type {
		case String:
			if option.Default == nil {
				option.Default = ""
			}
			option.Value = flagset.String(option.Name, option.Default.(string), "")
		case Bool:
			if option.Default == nil {
				option.Default = false
			}
			option.Value = flagset.Bool(option.Name, option.Default.(bool), "")
		case Int:
			if option.Default == nil {
				option.Default = 0
			}
			option.Value = flagset.Int(option.Name, option.Default.(int), "")
		}
	}

	e := flagset.Parse(args)
	if e != nil {
		return e
	}

	for i, argument := range t.getOrderedArguments() {
		switch argument.Type {
		case String:
			if argument.Default != nil && flagset.Arg(i) == "" {
				argument.Value = argument.Default
			} else if flagset.Arg(i) != "" {
				argument.Value = flagset.Arg(i)
			} else {
				return errInvalidParams
			}
		case Bool:
			boolVal, err := strconv.ParseBool(flagset.Arg(i))
			if (argument.Default != nil && flagset.Arg(i) == "") || (argument.Default != nil && err != nil) {
				argument.Value = argument.Default
			} else if flagset.Arg(i) != "" && err == nil {
				argument.Value = boolVal
			} else {
				return errInvalidParams
			}
		case Int:
			intVal, err := strconv.Atoi(flagset.Arg(i))
			if (argument.Default != nil && flagset.Arg(i) == "") || (argument.Default != nil && err != nil) {
				argument.Value = argument.Default
			} else if flagset.Arg(i) != "" && err == nil {
				argument.Value = intVal
			} else {
				return errInvalidParams
			}
		}
	}
	return nil
}
