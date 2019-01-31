package exec

import (
	"os"
	"strings"
)

var globalOpt = map[string]*Option{
	"help": &Option{
		Name:        "help",
		Type:        Bool,
		Default:     false,
		Description: "Display help",
	},
}

// Run is a high level function which adds special behaviour to the Tasks,
// namely displaying help to the user. If you wish to use the library without
// that feature use the FindTask function directly.
func run(rootTask *task) error {
	task, taskName, taskArgs := find(rootTask, os.Args)
	task.Options = mergeOptions(task.Options, globalOpt)
	return task.execute(taskName, taskArgs)
}

// FindTask attempts to recursively locate the Task which should be
// executed. The provided Task should be the root Task of the program
// containing all other subTasks. The array containing the provided
// arguments will most likely be the os.Args array. The function returns the
// located subTask, its name and the remaining unused arguments. Those
// values should be passed to the Task.Execute method.
func find(task *task, args []string) (*task, string, []string) {
	foundtask, foundArgs := findTask(task, args[1:])
	foundName := subTaskName(args, foundArgs)
	return foundtask, foundName, foundArgs
}

func findTask(task *task, args []string) (*task, []string) {
	for subtaskName, subtask := range task.subtasks {
		if len(args) > 0 && args[0] == subtaskName {
			return findTask(subtask, args[1:])
		}
	}
	return task, args
}

func subTaskName(originalArgs []string, remainingArgs []string) string {
	argOffset := len(originalArgs) - len(remainingArgs)
	return strings.Join(originalArgs[:argOffset], " ")
}
