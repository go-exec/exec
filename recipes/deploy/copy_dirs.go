package deploy

import (
	"fmt"
	"github.com/go-exec/exec"
)

func init() {
	exec := exec.Instance
	exec.Task("deploy:copy_dirs", func() {
		dirs := exec.Get("copy_dirs").Slice()

		for _, dir := range dirs {
			// Delete directory if exists.
			exec.Remote(fmt.Sprintf("if [ -d $(echo {{release_path}}/%s) ]; then rm -rf {{release_path}}/%s; fi", dir, dir))

			// Copy directory.
			exec.Remote(fmt.Sprintf("if [ -d $(echo {{deploy_path}}/current/%s) ]; then cp -rpf {{deploy_path}}/current/%s {{release_path}}%s; fi", dir, dir, dir))
		}
	}).ShortDescription("Copy directories")
}
