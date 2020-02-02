package deploy

import "github.com/go-exec/exec"

func init() {
	exec := exec.Instance
	exec.Task("deploy:lock", func() {
		locked := exec.Remote("if [ -f {{deploy_path}}/.dep/deploy.lock ]; then echo 'true'; fi").Bool()

		if locked {
			exec.Println("Deploy locked.\nRun deploy:unlock command to unlock.")
			return
		} else {
			exec.Remote("touch {{deploy_path}}/.dep/deploy.lock")
		}
	}).ShortDescription("Lock deploy")

	exec.Task("deploy:unlock", func() {
		exec.Remote("rm {{deploy_path}}/.dep/deploy.lock")
	}).ShortDescription("Unlock deploy")
}
