package php

import (
	"github.com/go-exec/exec"
	_ "github.com/go-exec/exec/recipes/deploy"
)

func init() {
	exec := exec.Instance
	exec.Set("http_user", false)
	exec.Set("http_group", false)

	exec.Set("composer_action", "install")
	exec.Set("composer_options", "{{composer_action}} --verbose --prefer-dist --no-progress --no-interaction --no-dev --optimize-autoloader")

	exec.Set("env_vars", "") // Variable assignment before cmds (for example, SYMFONY_ENV={{set}})

	exec.Set("bin/php", func() interface{} {
		return exec.Remote("which php").String()
	})

	exec.Set("bin/composer", func() interface{} {
		var composer string

		if exec.CommandExist("composer") {
			composer = exec.Remote("which composer").String()
		}

		if composer == "" {
			exec.Remote("cd {{release_path}} && curl -sS https://getcomposer.org/installer | {{bin/php}}")
			composer = "{{bin/php}} {{release_path}}/composer.phar"
		}

		return composer
	})

}
