package deploy

import (
	"fmt"
	"github.com/go-exec/exec"
)

func init() {
	exec.Task("cleanup", func() {
		releases := exec.Get("releases_list").Slice()
		keep := exec.Get("keep_releases").Int()

		if keep <= 0 || len(releases)+1 <= keep {
			exec.Println("No cleanup needed.")
			return
		} else {
			releases = releases[keep-1:]
		}

		// releases to be deleted, old ones
		for _, release := range releases {
			exec.Remote(fmt.Sprintf("rm -rf {{deploy_path}}/releases/%s", release))
		}

		exec.Remote("cd {{deploy_path}} && if [ -e release ]; then rm release; fi")
		exec.Remote("cd {{deploy_path}} && if [ -h release ]; then rm release; fi")
	}).ShortDescription("Cleaning up old releases")
}
