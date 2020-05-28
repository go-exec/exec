package exec

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServer_AddRole(t *testing.T) {
	s := &server{
		Name: "qa",
		Dsn:  "root@domain.com",
	}
	s.AddRole("qa").AddRole("test")

	require.Equal(t, s.roles, []string{"qa", "test"})
}

func TestServer_HasRole(t *testing.T) {
	s := &server{
		Name: "qa",
		Dsn:  "root@domain.com",
	}
	s.AddRole("qa")

	require.True(t, s.HasRole("qa"))
	require.False(t, s.HasRole("test"))
}

func TestServer_Set(t *testing.T) {
	s := &server{
		Name:    "qa",
		Dsn:     "root@domain.com",
		Configs: make(map[string]*config),
	}
	s.Set("config", "value")

	require.Contains(t, s.Configs, "config")
	require.Equal(t, s.Configs["config"].Value(), "value")
}

func TestServer_Key(t *testing.T) {
	s := &server{
		Name:      "qa",
		Dsn:       "root@domain.com",
		Configs:   make(map[string]*config),
		sshClient: &sshClient{},
	}
	s.Key("key")

	require.Contains(t, s.sshClient.keys, "key")
}

func TestServer_GetUser(t *testing.T) {
	s := &server{
		Name: "qa",
		Dsn:  "root@domain.com",
	}

	require.Equal(t, s.GetUser(), "root")
}

func TestServer_GetHost(t *testing.T) {
	s := &server{
		Name: "qa",
		Dsn:  "root@domain.com",
	}

	require.Equal(t, s.GetHost(), "domain.com")
}
