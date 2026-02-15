package auth

import (
	"context"
)

// Credential holds authentication data.
type Credential struct {
	Username string
	Password string
	Token    string
}

// CredentialProvider retrieves credentials for a specific server.
type CredentialProvider interface {
	Get(ctx context.Context, serverAddress string) (*Credential, error)
}
