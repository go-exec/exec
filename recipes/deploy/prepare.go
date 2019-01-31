package deploy

import "github.com/go-exec/exec"

func init() {
	exec.Task("deploy:prepare", func() {
		exec.Remote("if [ ! -d {{deploy_path}} ]; then mkdir -p {{deploy_path}}; fi")

		// Check for existing /current directory (not symlink)
		result := exec.Remote("if [ ! -L {{deploy_path}}/current ] && [ -d {{deploy_path}}/current ]; then echo true; fi").Bool()
		if result {
			exec.Println("There already is a directory (not symlink) named `current` in {{deploy_path}}. Remove this directory so it can be replaced with a symlink for atomic deployments.")
		}

		// Create metadata .dep dir.
		exec.Remote("cd {{deploy_path}} && if [ ! -d .dep ]; then mkdir .dep; fi")

		// Create releases dir.
		exec.Remote("cd {{deploy_path}} && if [ ! -d releases ]; then mkdir releases; fi")

		// Create shared dir.
		exec.Remote("cd {{deploy_path}} && if [ ! -d shared ]; then mkdir shared; fi")
	}).ShortDescription("Preparing server for deploy")
}
