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