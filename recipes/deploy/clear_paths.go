package deploy

import (
	"fmt"
	"github.com/go-exec/exec"
)

func init() {
	exec.Task("deploy:clear_paths", func() {
		paths := exec.Get("clear_paths").Slice()
		sudo := ""
		if exec.Get("clear_use_sudo").Bool() {
			sudo = "sudo"
		}

		for _, path := range paths {
			exec.Remote(fmt.Sprintf("%s rm -rf {{release_path}}/%s", sudo, path))
		}
	}).ShortDescription("Cleaning up files and/or directories")
}
