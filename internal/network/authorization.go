package network

import (
	"encoding/base64"
	"fmt"
)

type authorization interface {
	Authorization() string
}

// NewTokenAuthorization returns an authorization object capable of
// providing a Bearer Token authorization header for a request to UAA.
func NewTokenAuthorization(token string) tokenAuthorization {
	return tokenAuthorization(token)
}

type tokenAuthorization string

func (a tokenAuthorization) Authorization() string {
	return fmt.Sprintf("Bearer %s", a)
}

// NewBasicAuthorization returns an authorization object capable of
// providing a HTTP Basic authorization header for a request to UAA.
func NewBasicAuthorization(username, password string) basicAuthorization {
	return basicAuthorization{
		username: username,
		password: password,
	}
}

type basicAuthorization struct {
	username string
	password string
}

func (b basicAuthorization) Authorization() string {
	auth := b.username + ":" + b.password
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
}
