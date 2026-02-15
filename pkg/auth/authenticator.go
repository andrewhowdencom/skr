package auth

import (
	"encoding/base64"
	"net/http"
)

// Authenticator injects credentials into an HTTP request.
type Authenticator interface {
	Authenticate(req *http.Request) error
}

// AuthenticatorFromCredential creates an Authenticator based on the provided Credential.
func AuthenticatorFromCredential(cred *Credential) Authenticator {
	if cred.Token != "" {
		return &BearerAuthenticator{Token: cred.Token}
	}
	if cred.Username != "" && cred.Password != "" {
		return &BasicAuthenticator{Username: cred.Username, Password: cred.Password}
	}
	// Fallback or empty authenticator?
	// Returning nil might be safer, let caller handle it.
	return nil
}

// BasicAuthenticator adds Basic Auth headers.
type BasicAuthenticator struct {
	Username string
	Password string
}

func (a *BasicAuthenticator) Authenticate(req *http.Request) error {
	auth := a.Username + ":" + a.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encoded)
	return nil
}

// BearerAuthenticator adds Bearer Token headers.
type BearerAuthenticator struct {
	Token string
}

func (a *BearerAuthenticator) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return nil
}
