package deploy

import "github.com/go-exec/exec"

func init() {
	exec := exec.Instance
	exec.Task("deploy:symlink", func() {
		if exec.Remote("if [[ \"$(man mv 2>/dev/null)\" =~ '--no-target-directory' ]]; then echo 'true'; fi").Bool() {
			exec.Remote("mv -T {{deploy_path}}/release {{deploy_path}}/current")
		} else {
			// Atomic symlink does not supported.
			// Will use simple two steps switch.
			exec.Remote("cd {{deploy_path}} && {{bin/symlink}} {{release_path}} current") // Atomic override symlink.
			exec.Remote("cd {{deploy_path}} && rm release")                               // Remove release link.
		}
	}).ShortDescription("Creating symlink to release")
}
