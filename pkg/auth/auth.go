package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	ServiceName = "skr"
)

// Store implements credentials.Store
type Store struct{}

func NewStore() *Store {
	return &Store{}
}

// Get retrieves credentials from the keyring.
func (s *Store) Get(ctx context.Context, serverAddress string) (auth.Credential, error) {
	u, p, err := GetCredentials(serverAddress)
	if err != nil {
		return auth.Credential{}, err
	}
	return auth.Credential{
		Username: u,
		Password: p,
	}, nil
}

// Put stores credentials in the keyring.
func (s *Store) Put(ctx context.Context, serverAddress string, credential auth.Credential) error {
	return Login(serverAddress, credential.Username, credential.Password)
}

// Delete removes credentials from the keyring.
func (s *Store) Delete(ctx context.Context, serverAddress string) error {
	return Logout(serverAddress)
}

// AuthConfig holds a single credential.
type AuthConfig struct {
	Username string
	Password string
}

// Config is a map of registry -> AuthConfig.
// We still need to load *all* configs to find the right one if we scan.
// But keyring usually stores by (service, user).
// Standard pattern: Service="skr", User="<registry>".
// Password payload contains JSON with actual username/password?
// OR Service="skr", User="<registry>/<username>"?
// Docker cred helpers use the registry as the "ServerURL".
//
// For simplicity and standard keyring usage:
// Service: "skr"
// User: <registry_domain> (e.g. "ghcr.io")
// Password: <json_blob_of_creds> OR just the password/token?
// If we just store password, we lose the username.
// So we should store a JSON blob as the "Password"/Secret in the keyring.
type CredentialPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login stores credentials for a registry in the keyring.
func Login(registry, username, password string) error {
	payload := CredentialPayload{
		Username: username,
		Password: password,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal credential payload: %w", err)
	}

	// Service: skr
	// User: registry (e.g. ghcr.io)
	// Secret: JSON(user, pass)
	if err := keyring.Set(ServiceName, registry, string(data)); err != nil {
		return fmt.Errorf("failed to save credentials to keyring: %w", err)
	}
	return nil
}

// Logout removes credentials for a registry from the keyring.
func Logout(registry string) error {
	if err := keyring.Delete(ServiceName, registry); err != nil {
		if strings.Contains(err.Error(), "not found") { // naive check, error types vary by platform
			return fmt.Errorf("not logged in to %s", registry)
		}
		return fmt.Errorf("failed to delete credentials: %w", err)
	}
	return nil
}

// GetCredentials retrieves credentials for a registry.
func GetCredentials(registry string) (string, string, error) {
	data, err := keyring.Get(ServiceName, registry)
	if err != nil {
		return "", "", fmt.Errorf("failed to get credentials for %s: %w", registry, err)
	}

	var payload CredentialPayload
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return "", "", fmt.Errorf("failed to parse credential payload: %w", err)
	}

	return payload.Username, payload.Password, nil
}

// LoadConfig returns a map representation for compatibility,
// though iterating keyring items isn't always supported across all backends easily.
// For our use case (Push/Pull), we usually know the registry we are targeting.
// ORAS might want a resolver.
// The registry wrapper (pkg/registry) handles this.
// We don't strictly need to list all auths.
