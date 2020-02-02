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
	executor := exec.New()
	
	executor.Task("onStart", func() {
		executor.Set("startTime", time.Now())
	}).Private()

	executor.Task("onEnd", func() {
		executor.Println(fmt.Sprintf("Finished in %s!`", time.Now().Sub(executor.Get("startTime").Time()).String()))
	}).Private()

	type F struct {
		F func() interface{}
	}

	stage := executor.NewArgument("stage", "Provide the running stage")
	stage.Default = "qa"
	stage.Type = exec.String

	executor.AddArgument(stage)

	//run always on the server set by stage dynamically
	//executor.OnServers(func() []string {
	//	return []string{executor.GetArgument("stage").String()}
	//})

	arg2 := executor.NewArgument("arg2", "Provide the arg2")
	arg2.Default = "test"

	executor.AddArgument(arg2)

	executor.Set("env", "prod")

	executor.Set("bin/mysql", "mysql default")

	executor.Set("test", func() interface{} { return "text" })

	executor.Set("functest", F{F: func() interface{} {
		return "date"
	}})

	executor.Set("localUser", executor.Local("git config --get %s", "user.name"))

	executor.
		Server("prod1", "root@domain.com").
		AddRole("prod").
		Set("bin/mysql", "mysql prod")

	executor.
		Server("prod2", "root@domain.com").
		AddRole("prod").
		Set("bin/mysql", "mysql prod")

	executor.
		Server("qa", "root@domain.com").
		Key("~/.ssh/id_rsa").
		AddRole("qa").
		Set("bin/mysql", "mysql qa")

	executor.
		Server("stage", "root@domain.com").
		AddRole("stage")

	opt1 := executor.NewOption("opt1", "test")
	opt2 := executor.NewOption("opt2", "test")

	executor.
		Task("upload", func() {
			executor.Remote("ls -la /")
			executor.Upload("test.txt", "~/test.txt")
		})

	executor.
		Task("download", func() {
			executor.Remote("ls -la /")
			executor.Download("~/test.txt", "test.txt")
		})

	executor.
		Task("test1", func() {
			//fmt.Println(executor.TaskContext.GetOption("opt1").ToString())
			executor.Remote("echo Git user is: " + executor.Get("localUser").String())
			executor.Remote("ls -la /")
			fmt.Println(executor.TaskContext.GetArgument("stage"))
			fmt.Println(executor.TaskContext.GetArgument("arg2"))
		}).
		ShortDescription("Running test1 task").
		AddOption(opt1)

	executor.
		Task("test2", func() {
			executor.Remote("ls -la ~")
		}).
		ShortDescription("Running test2 task").
		AddOption(opt2)

	executor.
		Task("test3", func() {
			executor.Remote("ls -la ~/.ssh")
		}).
		ShortDescription("Running test3 task")

	//should avoid using RunLocal in a task that will run in a stage with multiple servers associated!
	executor.
		Task("local", func() {
			executor.Local("ls -la ~/Public; ls -la /Users/")
			executor.Local("docker")
		}).
		Once().
		ShortDescription("Running local task")

	executor.
		Task("yarn", func() {
			executor.Local("yarn")
		})

	executor.
		Task("docker", func() {
			executor.Local("docker stats")
		})

	executor.
		Task("docker-remote", func() {
			executor.Remote("docker")
		}).
		OnServers(func() []string {
			return []string{"prod1"}
		})

	executor.
		Task("get", func() {
			executor.Remote(fmt.Sprintf("%s", executor.Get("bin/mysql").String()))
			fmt.Println(executor.TaskContext.GetArgument("stage"))
			fmt.Println(executor.TaskContext.GetArgument("arg2"))
		}).
		ShortDescription("Testing get in different servers contexts")

	executor.
		Task("get2", func() {
			executor.Remote(fmt.Sprintf("%s", executor.Get("functest").Value().(F).F()))
		}).
		ShortDescription("Testing get2 in different servers contexts")

	executor.
		Task("get3", func() {
			executor.Println(executor.Get("test").String())
		}).
		ShortDescription("Testing get3 in different servers contexts").
		RemoveArgument("stage")

	executor.
		TaskGroup("deploy1", "test1", "test2").
		ShortDescription("Deploy code 1")

	executor.
		TaskGroup("deploy2", "local", "test3").
		ShortDescription("Deploy code 2")

	executor.
		TaskGroup("deploy3", "get").
		ShortDescription("Deploy code 3")

	executor.
		Task("onservers:a", func() {
			executor.RunIfNoBinary("docker", []string{
				"echo 'a'",
				"echo 'b'",
			})
		}).
		OnServers(func() []string {
			return []string{"prod1", "prod2"}
		})

	executor.
		Task("onservers:b", func() {
			executor.RunIfNoBinary("wget", []string{
				"echo 'a'",
				"echo 'b'",
			})
		}).
		OnServers(func() []string {
			return []string{"prod1", "prod2"}
		})

	executor.
		Task("onservers:c", func() {
			executor.RunIfNoBinary("docker", []string{
				"echo 'a'",
				"echo 'b'",
			})
		}).
		OnlyOnServers([]string{"prod1"})

	executor.
		Task("servercontext:host", func() {
			fmt.Println(executor.ServerContext.Name, executor.ServerContext.GetHost())
		}).
		OnServers(func() []string {
			return []string{"prod1", "prod2"}
		})

	executor.
		Task("onservers:read", func() {
			fmt.Printf("`%s`\n", executor.Remote("git config --get %s", "user.name").String())
		}).
		OnServers(func() []string {
			return []string{"prod1"}
		})

	executor.
		Task("ask", func() {
			response := executor.Ask("How are you?", "better")

			color.Yellow("Your response is `%s`", response)
		}).
		OnServers(func() []string {
			return []string{"prod1"}
		})

	executor.
		Task("ask-confirmation", func() {
			response := executor.AskWithConfirmation("Would you like to give it a shot?", true)

			color.Yellow("Your response is `%t`", response)

			response = executor.AskWithConfirmation("Would you like to give it a shot again?", false)

			color.Yellow("Your response is `%t`", response)
		}).
		OnServers(func() []string {
			return []string{"prod1"}
		})

	executor.
		Task("ask-choices", func() {
			response := executor.AskWithChoices("What are your choices?", map[string]interface{}{
				"default": []string{
					"agent",
				},
				"choices": []string{
					"agent",
					"tty",
					"ssh",
				},
			})

			color.Yellow("Your responses are `%v`", response)
		}).
		OnServers(func() []string {
			return []string{"prod1"}
		})

	executor.Before("test3", "local")
	executor.Before("get3", "test3")
	executor.Before("get3", "local")
	executor.After("local", "onservers:a")
	executor.After("local", "get3")
	executor.After("onservers:a", "local")

	executor.Init()
}
