package deploy

import "github.com/go-exec/exec"

func init() {
	exec := exec.Instance
	stage := exec.NewArgument("stage", "Provide the running stage")
	stage.Default = "qa"

	exec.AddArgument(stage)
}
