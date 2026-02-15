package auth

import (
	"fmt"
	"net/http"
)

// AuthTransport is an http.RoundTripper that injects credentials.
type AuthTransport struct {
	Base     http.RoundTripper
	Provider CredentialProvider
}

// NewAuthTransport creates a new AuthTransport.
func NewAuthTransport(base http.RoundTripper, provider CredentialProvider) *AuthTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &AuthTransport{
		Base:     base,
		Provider: provider,
	}
}

// RoundTrip implements http.RoundTripper.
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Attempt to get credentials for the target host
	cred, err := t.Provider.Get(req.Context(), req.URL.Host)
	if err == nil && cred != nil {
		auth := AuthenticatorFromCredential(cred)
		if auth != nil {
			if err := auth.Authenticate(req); err != nil {
				return nil, fmt.Errorf("failed to inject credentials: %w", err)
			}
		}
	}

	return t.Base.RoundTrip(req)
}
