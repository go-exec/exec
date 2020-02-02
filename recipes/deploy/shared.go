package deploy

import (
	"fmt"
	"github.com/go-exec/exec"
	"path"
	"strings"
)

func init() {
	exec := exec.Instance
	exec.Task("deploy:shared", func() {
		sharedPath := "{{deploy_path}}/shared"

		for _, dir := range exec.Get("shared_dirs").Slice() {
			// Create shared dir if it does not exist.
			exec.Remote(fmt.Sprintf("mkdir -p $sharedPath/%s", dir))

			// Copy shared dir files if they does not exist.
			exec.Remote(fmt.Sprintf("if [ -d $(echo {{release_path}}/%s) ]; then cp -rn {{release_path}}/%s %s; fi", dir, dir, sharedPath))

			// Remove from source.
			exec.Remote(fmt.Sprintf("if [ -d $(echo {{release_path}}/%s) ]; then rm -rf {{release_path}}/%s; fi", dir, dir))

			// Create path to shared dir in release dir if it does not exist.
			// (symlink will not create the path and will fail otherwise)
			exec.Remote(fmt.Sprintf("mkdir -p `dirname {{release_path}}/%s`", dir))

			// Symlink shared dir to release dir
			exec.Remote(fmt.Sprintf("{{bin/symlink}} %s/%s {{release_path}}/%s", sharedPath, dir, dir))
		}

		for _, file := range exec.Get("shared_files").Slice() {
			dir := path.Dir(file)
			i := strings.LastIndex(dir, "/")

			dirname := dir[i+1:]

			// Remove from source.
			exec.Remote(fmt.Sprintf("if [ -f $(echo {{release_path}}/%s) ]; then rm -rf {{release_path}}/%s; fi", file, file))

			// Ensure dir is available in release
			exec.Remote(fmt.Sprintf("if [ ! -d $(echo {{release_path}}/%s) ]; then mkdir -p {{release_path}}/%s;fi", dirname, dirname))

			// Create dir of shared file
			exec.Remote(fmt.Sprintf("mkdir -p %s/%s", sharedPath, dirname))

			// Touch shared
			exec.Remote(fmt.Sprintf("touch %s/%s", sharedPath, file))

			// Symlink shared dir to release dir
			exec.Remote(fmt.Sprintf("{{bin/symlink}} %s/%s {{release_path}}/%s", sharedPath, file, file))
		}
	}).ShortDescription("Creating symlinks for shared files and dirs")

}
