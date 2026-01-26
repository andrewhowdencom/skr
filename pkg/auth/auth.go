package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/zalando/go-keyring"
	"oras.land/oras-go/v2/registry/remote/auth"
)

const (
	ServiceName  = "skr"
	AuthFileName = "auth.json"
)

// Store implements credentials.Store
type Store struct{}

func NewStore() *Store {
	return &Store{}
}

// Get retrieves credentials from the keyring or fallback file.
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

// Put stores credentials in the keyring or fallback file.
func (s *Store) Put(ctx context.Context, serverAddress string, credential auth.Credential) error {
	return Login(serverAddress, credential.Username, credential.Password)
}

// Delete removes credentials.
func (s *Store) Delete(ctx context.Context, serverAddress string) error {
	return Logout(serverAddress)
}

// AuthConfig holds a single credential.
type AuthConfig struct {
	Username string
	Password string
}

type CredentialPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login stores credentials. Tries keyring first, falls back to file.
func Login(registry, username, password string) error {
	payload := CredentialPayload{
		Username: username,
		Password: password,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal credential payload: %w", err)
	}

	// Try keyring
	if err := keyring.Set(ServiceName, registry, string(data)); err != nil {
		// Log warning?
		// Fallback to file
		return saveToAuthFile(registry, payload)
	}
	return nil
}

// Logout removes credentials.
func Logout(registry string) error {
	// Try deleting from keyring
	errKeyring := keyring.Delete(ServiceName, registry)

	// Try deleting from file
	errFile := deleteFromAuthFile(registry)

	// If neither worked, return error (prioritizing keyring error if it's not just "not found")
	if errKeyring != nil && errFile != nil {
		// Simplification: just return one
		return fmt.Errorf("failed to logout (keyring: %v, file: %v)", errKeyring, errFile)
	}
	return nil
}

// GetCredentials retrieves credentials.
func GetCredentials(registry string) (string, string, error) {
	// 1. Try keyring
	data, err := keyring.Get(ServiceName, registry)
	if err == nil {
		var payload CredentialPayload
		if err := json.Unmarshal([]byte(data), &payload); err == nil {
			return payload.Username, payload.Password, nil
		}
	}

	// 2. Try auth file
	payload, err := getFromAuthFile(registry)
	if err == nil {
		return payload.Username, payload.Password, nil
	}

	return "", "", fmt.Errorf("credentials not found for %s", registry)
}

// File-based backup helpers

func getAuthFilePath() (string, error) {
	return xdg.ConfigFile(filepath.Join("skr", AuthFileName))
}

func loadAuthFile() (map[string]CredentialPayload, error) {
	path, err := getAuthFilePath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return make(map[string]CredentialPayload), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var auths map[string]CredentialPayload
	if err := json.Unmarshal(data, &auths); err != nil {
		return nil, err
	}
	return auths, nil
}

func saveToAuthFile(registry string, payload CredentialPayload) error {
	auths, err := loadAuthFile()
	if err != nil {
		return err
	}

	auths[registry] = payload

	data, err := json.MarshalIndent(auths, "", "  ")
	if err != nil {
		return err
	}

	path, err := getAuthFilePath()
	if err != nil {
		return err
	}

	// Ensure directory exists (xdg.ConfigFile might do this, but safe to check)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600) // Secure permissions
}

func deleteFromAuthFile(registry string) error {
	auths, err := loadAuthFile()
	if err != nil {
		return err
	}

	if _, ok := auths[registry]; !ok {
		return fmt.Errorf("not found in auth file")
	}

	delete(auths, registry)

	data, err := json.MarshalIndent(auths, "", "  ")
	if err != nil {
		return err
	}

	path, err := getAuthFilePath()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func getFromAuthFile(registry string) (CredentialPayload, error) {
	auths, err := loadAuthFile()
	if err != nil {
		return CredentialPayload{}, err
	}

	if payload, ok := auths[registry]; ok {
		return payload, nil
	}
	return CredentialPayload{}, fmt.Errorf("not found")
}
