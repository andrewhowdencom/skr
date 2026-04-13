package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestStore_Get_PopulatesRefreshToken(t *testing.T) {
	// Prepare a temporary XDG_CONFIG_HOME
	tmpDir, err := os.MkdirTemp("", "xdg-config-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)
	os.Setenv("XDG_CONFIG_HOME", tmpDir)
	xdg.Reload()
	defer xdg.Reload()

	// Write mock authentication.yaml for DefaultYAMLProvider
	skrConfigDir := filepath.Join(tmpDir, "skr")
	err = os.MkdirAll(skrConfigDir, 0755)
	require.NoError(t, err)

	authYAMLPath := filepath.Join(skrConfigDir, "authentication.yaml")
	
	// Create mock credentials that include a RefreshToken
	mockCreds := map[string]yamlCreds{
		"userpass.com": {
			Username:     "testuser",
			Password:     "testpass",
			RefreshToken: "testrefresh1",
		},
		"token.com": {
			Token:        "testtoken",
			RefreshToken: "testrefresh2",
		},
	}
	
	yamlData, err := yaml.Marshal(mockCreds)
	require.NoError(t, err)
	err = os.WriteFile(authYAMLPath, yamlData, 0644)
	require.NoError(t, err)

	store := NewStore()
	ctx := context.Background()

	cred1, err := store.Get(ctx, "userpass.com")
	require.NoError(t, err)

	assert.Equal(t, "testuser", cred1.Username)
	assert.Equal(t, "testpass", cred1.Password)
	assert.Equal(t, "", cred1.AccessToken)
	assert.Equal(t, "testrefresh1", cred1.RefreshToken)

	cred2, err := store.Get(ctx, "token.com")
	require.NoError(t, err)

	assert.Equal(t, "", cred2.Username)
	assert.Equal(t, "", cred2.Password)
	assert.Equal(t, "testtoken", cred2.AccessToken)
	assert.Equal(t, "testrefresh2", cred2.RefreshToken)
}
