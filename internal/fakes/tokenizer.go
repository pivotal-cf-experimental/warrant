package fakes

import (
	"encoding/json"

	"github.com/pivotal-golang/conceal"
)

type Tokenizer struct {
	cloak conceal.Cloak
}

func NewTokenizer(key string) Tokenizer {
	cloak, err := conceal.NewCloak([]byte(key))
	if err != nil {
		panic(err)
	}

	return Tokenizer{
		cloak: cloak,
	}
}

func (t Tokenizer) Encrypt(token Token) string {
	tokenData, err := json.Marshal(token)
	if err != nil {
		panic(err)
	}

	encrypted, err := t.cloak.Veil(tokenData)
	if err != nil {
		panic(err)
	}

	return string(encrypted)
}

func (t Tokenizer) Decrypt(encryptedToken string) Token {
	tokenData, err := t.cloak.Unveil([]byte(encryptedToken))
	if err != nil {
		panic(err)
	}

	var token Token
	err = json.Unmarshal(tokenData, &token)
	if err != nil {
		panic(err)
	}

	return token
}

func (t Tokenizer) Validate(token, expected Token) bool {
	if ok := token.HasAudiences(expected.Audiences); !ok {
		return false
	}

	if ok := token.HasScopes(expected.Scopes); !ok {
		return false
	}

	return true
}

type Token struct {
	Scopes    []string
	Audiences []string
	UserID    string
}

func (t Token) HasScopes(scopes []string) bool {
	for _, scope := range scopes {
		if !contains(t.Scopes, scope) {
			return false
		}
	}
	return true
}

func (t Token) HasAudiences(audiences []string) bool {
	for _, audience := range audiences {
		if !contains(t.Audiences, audience) {
			return false
		}
	}
	return true
}

func contains(collection []string, item string) bool {
	for _, elem := range collection {
		if elem == item {
			return true
		}
	}

	return false
}
