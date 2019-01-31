package php

import "github.com/go-exec/exec"

func init() {
	exec.Task("deploy:vendors", func() {
		exec.Remote("cd {{release_path}} && {{env_vars}} {{bin/composer}} {{composer_options}}")
	}).ShortDescription("Installing vendors")
}
