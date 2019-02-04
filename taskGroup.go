package exec

type taskGroup struct {
	Name  string
	tasks []string
	task  *task
}

func (t *taskGroup) ShortDescription(desc string) *taskGroup {
	t.task.shortDescription = desc
	return t
}

func (t *taskGroup) AddArgument(argument *Argument) *taskGroup {
	t.task.Arguments[argument.Name] = argument
	return t
}

func (t *taskGroup) AddOption(option *Option) *taskGroup {
	t.task.Options[option.Name] = option
	return t
}

func (t *taskGroup) OnServers(f func() []string) {
	t.task.serverContextF = f
}
