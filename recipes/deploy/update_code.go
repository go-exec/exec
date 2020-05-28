package deploy

import (
	"fmt"
	"github.com/go-exec/exec"
)

func init() {
	exec := exec.Instance
	exec.Task("deploy:update_code", func() {
		repository := exec.Get("repository").String()
		branch := exec.Get("branch").String()
		tag := exec.Get("tag").String()
		git := exec.Get("bin/git").String()
		gitCache := exec.Get("git_cache").Bool()
		depth := ""
		at := ""
		if !gitCache {
			depth = "--depth 1"
		}

		// If option `branch` is set.
		if exec.TaskContext.HasOption("branch") {
			inputBranch := exec.TaskContext.GetOption("branch").String()
			if inputBranch != "" {
				branch = inputBranch
			}
		}

		// Branch may come from option or from configuration.
		if branch != "" {
			at = "-b " + branch
		}

		// If option `tag` is set
		if exec.TaskContext.HasOption("tag") {
			inputTag := exec.TaskContext.GetOption("tag").String()
			if inputTag != "" {
				tag = inputTag
			}
		}

		// Tag may come from option or from configuration.
		if tag != "" {
			at = "-b " + tag
		}

		// If option `tag` is not set and option `revision` is set
		revision := ""
		if tag == "" && exec.TaskContext.HasOption("revision") {
			revision = exec.TaskContext.GetOption("revision").String()
			if revision != "" {
				depth = ""
			}
		}

		releases := exec.Get("releases_list").Slice()

		if gitCache && len(releases) > 0 {
			if exec.Remote(fmt.Sprintf("%s clone %s --recursive -q --reference {{deploy_path}}/releases/%s --dissociate %s {{release_path}}", git, at, releases[0], repository)).HasError() {
				// If {{deploy_path}}/releases/{$releases[0]} has a failed git clone, is empty, shallow etc, git would throw error and give up. So we're forcing it to act without reference in this situation
				exec.Remote(fmt.Sprintf("%s clone %s --recursive -q %s {{release_path}}", git, at, repository))
			}
		} else {
			// if we're using git cache this would be identical to above code in catch - full clone. If not, it would create shallow clone.
			exec.Remote(fmt.Sprintf("%s clone %s %s --recursive -q %s {{release_path}}", git, at, depth, repository))
		}

		if revision != "" {
			exec.Remote(fmt.Sprintf("cd {{release_path}} && %s checkout %s", git, revision))
		}
	}).ShortDescription("Update code")
}
