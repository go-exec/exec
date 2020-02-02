package symfony

import (
	"github.com/go-exec/exec"
	_ "github.com/go-exec/exec/recipes/php"
)

func init() {
	exec := exec.Instance
	exec.
		TaskGroup(
			"deploy",
			"deploy:prepare",
			"deploy:lock",
			"deploy:release",
			"deploy:update_code",
			"deploy:clear_paths",
			//"deploy:create_cache_dir",
			"deploy:shared",
			//"deploy:assets",
			//"deploy:vendors",
			//"deploy:assets:install",
			//"deploy:assetic:dump",
			//"deploy:cache:warmup",
			//"deploy:writable",
			"deploy:symlink",
			"deploy:unlock",
			"cleanup",
		).
		ShortDescription("Deploy code")
}
