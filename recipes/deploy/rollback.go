package deploy

import (
	"fmt"
	"github.com/go-exec/exec"
)

func init() {
	exec := exec.Instance
	exec.Task("rollback", func() {
		releases := exec.Get("releases_list").Slice()

		if len(releases) > 1 {
			// Symlink to old release.
			exec.Remote(fmt.Sprintf("cd {{deploy_path}} && {{bin/symlink}} {{deploy_path}}/releases/%s current", releases[1]))

			// Remove release
			exec.Remote(fmt.Sprintf("rm -rf {{deploy_path}}/releases/%s", releases[0]))

			exec.Println(fmt.Sprintf("Rollback to `%s` release was successful.", releases[1]))
		}
	}).ShortDescription("Rollback to previous release")

}
