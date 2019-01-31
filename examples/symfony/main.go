package main

import (
	"github.com/go-exec/exec"
	_ "github.com/go-exec/exec/recipes/symfony"
)

/*
Example of deploying a Symfony app using the deploy recipes
*/
func main() {
	exec.Set("repository", "git@github.com:namespace/app.git")
	exec.Set("shared_files", []string{})
	exec.Set("shared_dirs", []string{"var/logs", "vendor", "web/uploads", "web/media", "node_modules"})
	exec.Set("writable_dirs", []string{"var/cache", "var/logs", "../../shared/web/uploads", "../../shared/web/media"})
	exec.Set("http_user", "www-data")
	exec.Set("use_relative_symlink", false)
	exec.Set("deploy_path", "/var/www/{{domain}}")
	exec.Set("env_vars", "SYMFONY_ENV={{env}}")

	//exec.GetArgument("stage").Default = "qa"

	exec.
		Server("qa", "root@qa.domain.com").
		AddRole("qa").
		Set("domain", "qa.domain.com").
		Set("env", "prod").
		Set("branch", "master")

	exec.
		Server("prod", "root@domain.com").
		AddRole("prod").
		Set("domain", "domain.com").
		Set("env", "qa").
		Set("branch", "production")

	//run always on the server set by stage dynamically
	exec.OnServer(func() string {
		return exec.GetArgument("stage").String()
	})

	exec.Init()
}
