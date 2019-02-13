package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/go-exec/exec"
	"time"
)

/*
Example with general setup of tasks
*/
func main() {
	exec.Task("onStart", func() {
		exec.Set("startTime", time.Now())
	}).Private()

	exec.Task("onEnd", func() {
		exec.Println(fmt.Sprintf("Finished in %s!`", time.Now().Sub(exec.Get("startTime").Time()).String()))
	}).Private()

	type F struct {
		F func() interface{}
	}

	stage := exec.NewArgument("stage", "Provide the running stage")
	stage.Default = "qa"
	stage.Type = exec.String

	exec.AddArgument(stage)

	//run always on the server set by stage dynamically
	//exec.OnServers(func() []string {
	//	return []string{exec.GetArgument("stage").String()}
	//})

	arg2 := exec.NewArgument("arg2", "Provide the arg2")
	arg2.Default = "test"

	exec.AddArgument(arg2)

	exec.Set("env", "prod")

	exec.Set("bin/mysql", "mysql default")

	exec.Set("test", func() interface{} { return "text" })

	exec.Set("functest", F{F: func() interface{} {
		return "date"
	}})

	exec.Set("localUser", exec.Local("git config --get %s", "user.name"))

	exec.
		Server("prod1", "root@domain.com").
		AddRole("prod").
		Set("bin/mysql", "mysql prod")

	exec.
		Server("prod2", "root@domain.com").
		AddRole("prod").
		Set("bin/mysql", "mysql prod")

	exec.
		Server("qa", "root@domain.com").
		Key("~/.ssh/id_rsa").
		AddRole("qa").
		Set("bin/mysql", "mysql qa")

	exec.
		Server("stage", "root@domain.com").
		AddRole("stage")

	opt1 := exec.NewOption("opt1", "test")
	opt2 := exec.NewOption("opt2", "test")

	exec.
		Task("upload", func() {
			exec.Remote("ls -la /")
			exec.Upload("test.txt", "~/test.txt")
		})

	exec.
		Task("download", func() {
			exec.Remote("ls -la /")
			exec.Download("~/test.txt", "test.txt")
		})

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
		Once().
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
		TaskGroup("deploy1", "test1", "test2").
		ShortDescription("Deploy code 1")

	exec.
		TaskGroup("deploy2", "local", "test3").
		ShortDescription("Deploy code 2")

	exec.
		TaskGroup("deploy3", "get").
		ShortDescription("Deploy code 3")

	exec.
		Task("onservers:a", func() {
			exec.RunIfNoBinary("docker", []string{
				"echo 'a'",
				"echo 'b'",
			})
		}).
		OnServers(func() []string {
			return []string{"prod1", "prod2"}
		})

	exec.
		Task("onservers:b", func() {
			exec.RunIfNoBinary("wget", []string{
				"echo 'a'",
				"echo 'b'",
			})
		}).
		OnServers(func() []string {
			return []string{"prod1", "prod2"}
		})

	exec.
		Task("onservers:c", func() {
			exec.RunIfNoBinary("docker", []string{
				"echo 'a'",
				"echo 'b'",
			})
		}).
		OnlyOnServers([]string{"prod1"})

	exec.
		Task("servercontext:host", func() {
			fmt.Println(exec.ServerContext.Name, exec.ServerContext.GetHost())
		}).
		OnServers(func() []string {
			return []string{"prod1", "prod2"}
		})

	exec.
		Task("onservers:read", func() {
			fmt.Printf("`%s`\n", exec.Remote("git config --get %s", "user.name").String())
		}).
		OnServers(func() []string {
			return []string{"prod1"}
		})

	exec.
		Task("ask", func() {
			response := exec.Ask("How are you?", map[string]string{
				"default": "better",
			})

			color.Yellow("Your response is `%s`", response)
		}).
		OnServers(func() []string {
			return []string{"prod1"}
		})

	exec.Before("test3", "local")
	exec.Before("get3", "test3")
	exec.Before("get3", "local")
	exec.After("local", "onservers:a")
	exec.After("local", "get3")
	exec.After("onservers:a", "local")

	exec.Init()
}
