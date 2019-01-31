package exec

type taskGroup struct {
	Name          string
	tasks         []string
	task          *task
	onlyOnServers []string
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

func (t *taskGroup) OnlyOnServers(servers []string) *taskGroup {
	t.onlyOnServers = servers
	return t
}

func (t *taskGroup) OnServer(f func() string) {
	t.task.serverContextF = f
}
