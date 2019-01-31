package main

import (
	"fmt"
	"github.com/go-exec/exec"
)

/*
Example with general setup of tasks
*/
func main() {
	type F struct {
		F func() interface{}
	}

	stage := exec.NewArgument("stage", "Provide the running stage")
	stage.Default = "qa"
	stage.Type = exec.String

	exec.AddArgument(stage)

	//run always on the server set by stage dynamically
	exec.OnServer(func() string {
		return exec.GetArgument("stage").String()
	})

	arg2 := exec.NewArgument("arg2", "Provide the arg2")
	arg2.Default = "test"

	exec.AddArgument(arg2)

	exec.Set("env", "prod")

	exec.Set("bin/mysql", "mysql default")

	exec.Set("test", func() interface{} { return "text" })

	exec.Set("functest", F{F: func() interface{} {
		return "date"
	}})

	exec.Set("localUser", exec.Local("git config --get user.name"))

	exec.
		Server("prod1", "root@localhost").
		AddRole("prod").
		Set("bin/mysql", "mysql prod")

	exec.
		Server("prod2", "root@localhost").
		AddRole("prod").
		Set("bin/mysql", "mysql prod")

	exec.
		Server("qa", "root@localhost").
		Key("~/.ssh/id_rsa").
		AddRole("qa").
		Set("bin/mysql", "mysql qa")

	exec.
		Server("stage", "root@localhost").
		AddRole("stage")

	opt1 := exec.NewOption("opt1", "test")
	opt2 := exec.NewOption("opt2", "test")

	exec.
		Task("test1", func() {
			//fmt.Println(exec.TaskContext.GetOption("opt1").ToString())
			exec.Remote("echo Git user is: " + exec.Get("localUser").String())
			exec.Remote("ls -la /")
			fmt.Println(exec.TaskContext.GetArgument("stage"))
			fmt.Println(exec.TaskContext.GetArgument("arg2"))
		}).
		ShortDescription("Running test1 task").
		AddOption(opt1)

	exec.
		Task("test2", func() {
			exec.Remote("ls -la ~")
		}).
		ShortDescription("Running test2 task").
		AddOption(opt2)

	exec.
		Task("test3", func() {
			exec.Remote("ls -la ~/.ssh")
		}).
		ShortDescription("Running test3 task")

	//should avoid using RunLocal in a task that will run in a stage with multiple servers associated!
	exec.
		Task("local", func() {
			exec.Local("ls -la ~/Public; ls -la /Users/")
		}).
		ShortDescription("Running local task")

	exec.
		Task("get", func() {
			exec.Remote(fmt.Sprintf("%s", exec.Get("bin/mysql").String()))
			fmt.Println(exec.TaskContext.GetArgument("stage"))
			fmt.Println(exec.TaskContext.GetArgument("arg2"))
		}).
		ShortDescription("Testing get in different servers contexts")

	exec.
		Task("get2", func() {
			exec.Remote(fmt.Sprintf("%s", exec.Get("functest").Value().(F).F()))
		}).
		ShortDescription("Testing get2 in different servers contexts")

	exec.
		Task("get3", func() {
			exec.Println(exec.Get("test").String())
		}).
		ShortDescription("Testing get3 in different servers contexts")

	exec.
		TaskGroup("deploy", "test1", "test2").
		ShortDescription("Deploy code")

	exec.Init()
}
