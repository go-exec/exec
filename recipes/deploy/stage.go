package deploy

import "github.com/go-exec/exec"

func init() {
	stage := exec.NewArgument("stage", "Provide the running stage")
	stage.Default = "qa"

	exec.AddArgument(stage)
}
