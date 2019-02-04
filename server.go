package exec

import "strings"

type server struct {
	Name    string
	Dsn     string
	Configs map[string]*config
	key     *string

	roles     []string
	sshClient *sshClient
}

func (s *server) AddRole(role string) *server {
	s.roles = append(s.roles, role)
	return s
}

func (s *server) HasRole(role string) bool {
	for _, r := range s.roles {
		if role == r {
			return true
		}
	}
	return false
}

func (s *server) Set(name string, value interface{}) *server {
	s.Configs[name] = &config{Name: name, value: value}
	return s
}

func (s *server) Key(file string) *server {
	s.key = &file
	s.sshClient.keys = append(s.sshClient.keys, file)
	return s
}

func (s *server) GetUser() string {
	return s.Dsn[:strings.Index(s.Dsn, "@")-1]
}

func (s *server) GetHost() string {
	if strings.Index(s.Dsn, ":") > 0 {
		return s.Dsn[strings.Index(s.Dsn, "@")+1 : strings.Index(s.Dsn, ":")]
	}
	return s.Dsn[strings.Index(s.Dsn, "@")+1:]
}
