package deploy

import (
	"fmt"
	"github.com/go-exec/exec"
	"log"
	"strings"
)

func init() {
	exec := exec.Instance
	exec.Task("deploy:writable", func() {
		dirs := strings.Join(exec.Get("writable_dirs").Slice(), " ")
		mode := exec.Get("writable_mode").String()
		sudo := ""
		if exec.Get("writable_use_sudo").Bool() {
			sudo = "sudo"
		}

		httpUser := exec.Get("http_user").String()

		if dirs == "" {
			return
		}

		if httpUser == "" && mode != "chmod" {
			// Detect http user in process list.
			httpUser = exec.Remote("ps axo user,comm | grep -E '[a]pache|[h]ttpd|[_]www|[w]ww-data|[n]ginx' | grep -v root | head -1 | cut -d\\  -f1").String()

			if httpUser == "" {
				log.Panicln("Can't detect http user name.\n Please setup `http_user` config parameter.")
			}
		}

		//try {
		exec.Cd("{{release_path}}")

		if mode == "chown" {
			// Change owner.
			// -R   operate on files and directories recursively
			// -L   traverse every symbolic link to a directory encountered
			exec.Remote(fmt.Sprintf("%s chown -RL %s %s", sudo, httpUser, dirs))
		} else if mode == "chgrp" {
			// Change group ownership.
			// -R   operate on files and directories recursively
			// -L   if a command line argument is a symbolic link to a directory, traverse it
			httpGroup := exec.Get("http_group").String()
			if httpGroup == "" {
				log.Panicln("Please setup `http_group` config parameter.")
			}
			exec.Remote(fmt.Sprintf("%s chgrp -RH %s %s", sudo, httpGroup, dirs))
		} else if mode == "chmod" {
			exec.Remote(fmt.Sprintf("%s chmod -R {{writable_chmod_mode}} %s", sudo, dirs))
		} else if mode == "acl" {
			if strings.Contains(exec.Remote("chmod 2>&1; true").String(), "+a") {
				// Try OS-X specific setting of access-rights

				exec.Remote(fmt.Sprintf("%s chmod +a \"%s allow delete,write,append,file_inherit,directory_inherit\" %s", sudo, httpUser, dirs))
				exec.Remote(fmt.Sprintf("%s chmod +a \"`whoami` allow delete,write,append,file_inherit,directory_inherit\" %s", sudo, dirs))
			} else if exec.CommandExist("setfacl") {
				if sudo != "" {
					exec.Remote(fmt.Sprintf("%s setfacl -R -m u:\"%s\":rwX -m u:`whoami`:rwX %s", sudo, httpUser, dirs))
					exec.Remote(fmt.Sprintf("%s setfacl -dR -m u:\"%s\":rwX -m u:`whoami`:rwX %s", sudo, httpUser, dirs))
				} else {
					// When running without sudo, exception may be thrown
					// if executing setfacl on files created by http user (in directory that has been setfacl before).
					// These directories/files should be skipped.
					// Now, we will check each directory for ACL and only setfacl for which has not been set before.
					writeableDirs := exec.Get("writable_dirs").Slice()
					for _, dir := range writeableDirs {
						// Check if ACL has been set or not
						hasfacl := exec.Remote(fmt.Sprintf("getfacl -p %s | grep \"^user:%s:.*w\" | wc -l", dir, httpUser)).Bool()
						// Set ACL for directory if it has not been set before
						if !hasfacl {
							exec.Remote(fmt.Sprintf("setfacl -R -m u:\"%s\":rwX -m u:`whoami`:rwX %s", httpUser, dir))
							exec.Remote(fmt.Sprintf("setfacl -dR -m u:\"%s\":rwX -m u:`whoami`:rwX %s", httpUser, dir))
						}
					}
				}
			} else {
				log.Panicln("Cant't set writable dirs with ACL.")
			}
		} else {
			log.Panicln("Unknown writable_mode `$mode`.")
		}
		/*} catch (\RuntimeException $e) {
		  $formatter = exec::get()->getHelper('formatter');

		  $errorMessage = [
		  "Unable to setup correct permissions for writable dirs.                  ",
		  "You need to configure sudo's sudoers files to not prompt for password,",
		  "or setup correct permissions manually.                                  ",
		  ];
		  write($formatter->formatBlock($errorMessage, 'error', true));

		  throw $e;
		  }*/
	}).ShortDescription("Make writable dirs")

}
