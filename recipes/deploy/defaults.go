package deploy

import (
	"fmt"
	e "github.com/go-exec/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func init() {
	exec := e.Instance
	exec.Set("keep_releases", 5)

	exec.Set("repository", "") // Repository to deploy.
	exec.Set("branch", "")     // Branch to deploy.
	exec.Set("tag", "")        // Tag to deploy.

	exec.Set("shared_dirs", []string{})
	exec.Set("shared_files", []string{})

	exec.Set("copy_dirs", []string{})

	exec.Set("writable_dirs", []string{})
	exec.Set("writable_mode", "acl")        // chmod, chown, chgrp or acl.
	exec.Set("writable_use_sudo", false)    // Using sudo in writable commands?
	exec.Set("writable_chmod_mode", "0755") // For chmod mode

	exec.Set("clear_paths", []string{}) // Relative path from deploy_path
	exec.Set("clear_use_sudo", false)   // Using sudo in clean commands?

	exec.Set("use_relative_symlink", true)

	exec.Set("git_cache", func() interface{} { //whether to use git cache - faster cloning by borrowing objects from existing clones.
		gitVersion := exec.Remote("{{bin/git}} version").String()
		version := "1.0.0"

		re := regexp.MustCompile(`((\d+\.?)+)`)
		if re.MatchString(gitVersion) {
			version = re.FindStringSubmatch(gitVersion)[0]
		}

		versionS := strings.Split(version, ".")

		i1, _ := strconv.Atoi(versionS[0])
		i2, _ := strconv.Atoi(versionS[1])

		if i1 >= 2 && i2 >= 3 {
			return true
		} else {
			return false
		}
	})

	exec.Set("bin/git", func() interface{} {
		return exec.Remote("which git").String()
	})

	exec.Set("bin/symlink", func() interface{} {
		if exec.Get("use_relative_symlink").Bool() {
			// Check if target system supports relative symlink.
			if exec.Remote("if [[ \"$(man ln 2>/dev/null)\" =~ \"--relative\" ]]; then echo 'true'; fi").Bool() {
				return "ln -nfs --relative"
			}
		}
		return "ln -nfs"
	})

	branch := exec.NewOption("branch", "Branch to deploy")
	branch.Type = e.String
	exec.AddOption(branch)

	tag := exec.NewOption("tag", "Tag to deploy")
	tag.Type = e.String
	exec.AddOption(tag)

	revision := exec.NewOption("revision", "Revision to deploy")
	revision.Type = e.String
	exec.AddOption(revision)

	exec.Task("current", func() {
		exec.Println("Current release: {{current_path}}")
	})

	/**
	 * Success message
	 */
	exec.Task("success", func() {
		exec.Println("Successfully deployed!")
	}).Private()

	/**
	 * Deploy failure
	 */
	exec.Task("deploy:failed", func() {
	}).Private()

	exec.Task("onStart", func() {
		exec.Println("Start")
		exec.Set("startTime", time.Now())
	}).Once().Private()

	exec.Task("onEnd", func() {
		exec.Println(fmt.Sprintf("Finished in %s!", time.Since(exec.Get("startTime").Time()).String()))
		exec.Println("End")
	}).Once().Private()
}
