package auth

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYAMLFileProvider(t *testing.T) {
	// Create a temporary YAML file
	content := `
ghcr.io:
  username: "testuser"
  password: "testpassword"
docker.io:
  token: "testtoken"
`
	tmpdb, err := os.CreateTemp("", "auth-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpdb.Name())

	_, err = tmpdb.WriteString(content)
	require.NoError(t, err)
	tmpdb.Close()

	provider := &YAMLFileProvider{Path: tmpdb.Name()}
	ctx := context.Background()

	t.Run("Basic Auth", func(t *testing.T) {
		cred, err := provider.Get(ctx, "ghcr.io")
		require.NoError(t, err)
		require.NotNil(t, cred)

		assert.Equal(t, "testuser", cred.Username)
		assert.Equal(t, "testpassword", cred.Password)

		// Verify Authenticator creation
		auth := AuthenticatorFromCredential(cred)
		req, _ := http.NewRequest("GET", "http://ghcr.io", nil)
		err = auth.Authenticate(req)
		require.NoError(t, err)

		username, password, ok := req.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "testuser", username)
		assert.Equal(t, "testpassword", password)
	})

	t.Run("Bearer Token", func(t *testing.T) {
		cred, err := provider.Get(ctx, "docker.io")
		require.NoError(t, err)
		require.NotNil(t, cred)

		assert.Equal(t, "testtoken", cred.Token)

		// Verify Authenticator creation
		auth := AuthenticatorFromCredential(cred)
		req, _ := http.NewRequest("GET", "http://docker.io", nil)
		err = auth.Authenticate(req)
		require.NoError(t, err)

		header := req.Header.Get("Authorization")
		assert.Equal(t, "Bearer testtoken", header)
	})

	t.Run("Not Found", func(t *testing.T) {
		_, err := provider.Get(ctx, "example.com")
		assert.Error(t, err)
	})
}
