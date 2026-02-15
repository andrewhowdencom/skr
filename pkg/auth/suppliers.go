package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

// --- Default YAML Provider ---

type DefaultYAMLProvider struct{}

func (s *DefaultYAMLProvider) Get(ctx context.Context, serverAddress string) (*Credential, error) {
	// Look for ~/.config/skr/authentication.yaml (or similar XDG path)
	path, err := xdg.ConfigFile(filepath.Join("skr", "authentication.yaml"))
	if err != nil {
		return nil, err
	}

	// Reuse YAMLFileProvider logic
	delegate := &YAMLFileProvider{Path: path}
	return delegate.Get(ctx, serverAddress)
}

// --- YAML File Provider ---

type YAMLFileProvider struct {
	Path string
}

type yamlCreds struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
}

func (s *YAMLFileProvider) Get(ctx context.Context, serverAddress string) (*Credential, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}

	var credsMap map[string]yamlCreds
	if err := yaml.Unmarshal(data, &credsMap); err != nil {
		return nil, fmt.Errorf("failed to parse auth file: %w", err)
	}

	creds, ok := credsMap[serverAddress]
	if !ok {
		return nil, fmt.Errorf("no credentials found for %s in %s", serverAddress, s.Path)
	}

	// Prefer Token if available, else User/Pass
	if creds.Token != "" {
		return &Credential{Token: creds.Token}, nil
	}

	if creds.Username != "" && creds.Password != "" {
		return &Credential{Username: creds.Username, Password: creds.Password}, nil
	}

	return nil, fmt.Errorf("incomplete credentials for %s", serverAddress)
}

// --- Keyring Provider ---

type KeyringProvider struct{}

func (s *KeyringProvider) Get(ctx context.Context, serverAddress string) (*Credential, error) {
	data, err := keyring.Get(ServiceName, serverAddress)
	if err != nil {
		return nil, err
	}

	var payload CredentialPayload
	if err := json.Unmarshal([]byte(data), &payload); err != nil {
		return nil, fmt.Errorf("failed to parse keyring data: %w", err)
	}

	return &Credential{Username: payload.Username, Password: payload.Password}, nil
}

// --- Legacy JSON Provider ---

type LegacyJSONProvider struct{}

func (s *LegacyJSONProvider) Get(ctx context.Context, serverAddress string) (*Credential, error) {
	payload, err := getFromAuthFile(serverAddress) // Using existing helper in auth.go
	if err != nil {
		return nil, err
	}
	return &Credential{Username: payload.Username, Password: payload.Password}, nil
}

// --- Chain Provider ---

type ChainProvider struct {
	Providers []CredentialProvider
}

func (s *ChainProvider) Get(ctx context.Context, serverAddress string) (*Credential, error) {
	var lastErr error
	for _, provider := range s.Providers {
		cred, err := provider.Get(ctx, serverAddress)
		if err == nil {
			return cred, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
