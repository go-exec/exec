package deploy

import (
	"fmt"
	"github.com/go-exec/exec"
	"regexp"
	"strconv"
	"strings"
)

func init() {
	exec.Set("keep_releases", -1)

	exec.Set("release_name", func() interface{} {
		list := exec.Get("releases_list").Slice()
		var max int = 0

		for _, lv := range list {
			if value, err := strconv.Atoi(lv); err == nil && value > max {
				max = value
			}
		}

		return strconv.Itoa(max + 1)
	})

	/**
	 * Return list of releases on server.
	 */
	exec.Set("releases_list", func() interface{} {
		exec.Cd("{{deploy_path}}")

		// If there is no releases return empty list.
		if !exec.Remote("[ -d releases ] && [ \"$(ls -A releases)\" ] && echo \"true\" || echo \"false\"").Bool() {
			return []string{}
		}

		// Will list only dirs in releases.
		re := regexp.MustCompile(`[\d]+/`)
		list := exec.Remote("cd releases && ls -t -d */ -1").Slice("\n")
		for k, lv := range list {
			lv = strings.TrimSpace(lv)
			if re.MatchString(lv) {
				list[k] = strings.TrimRight(lv, "/")
			}
		}

		releases := []string{} // Releases list.

		// Collect releases based on .dep/releases info.
		// Other will be ignored.
		if exec.Remote("if [ -f .dep/releases ]; then echo \"true\"; fi").Bool() {
			meta := exec.Remote("cat .dep/releases").Slice("\n")

			for _, lv := range list {
				for _, mv := range meta {
					vs := strings.Split(strings.TrimSpace(mv), ",")
					release := vs[1]

					if lv == release {
						releases = append(releases, release)
					}
				}
			}
		}

		return releases
	})

	exec.Set("release_path", func() interface{} {
		if !exec.Remote("if [ -h {{deploy_path}}/release ]; then echo 'true'; fi").Bool() {
			exec.Println("Release path does not found.\n" +
				"Run deploy:release to create a new release.")
			return nil
		}

		link := exec.Remote("readlink {{deploy_path}}/release").String()

		if strings.HasPrefix(link, "/") {
			return link
		} else {
			return exec.Get("deploy_path").String() + "/" + link
		}
	})

	exec.Task("deploy:release", func() {
		exec.Cd("{{deploy_path}}")

		// Clean up if there is unfinished release.
		if exec.Remote("if [ -h release ]; then echo 'true'; fi").Bool() {
			exec.Remote("rm -rf \"$(readlink release)\"") // Delete release.
			exec.Remote("rm release")                     // Delete symlink.
		}

		releasePath := exec.Parse("{{deploy_path}}/releases/{{release_name}}")

		// Metainfo.
		// Save metainfo about release.
		exec.Remote("echo `date +'%Y%m%d%H%M%S'`,{{release_name}} >> .dep/releases")

		// Make new release.
		exec.Remote(fmt.Sprintf("mkdir %s", releasePath))
		exec.Remote(fmt.Sprintf("{{bin/symlink}} %s {{deploy_path}}/release", releasePath))
	})
}
