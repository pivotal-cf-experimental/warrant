package server

import (
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type tokenizer struct {
	key []byte
}

func newTokenizer(key string) tokenizer {
	return tokenizer{
		key: []byte(key),
	}
}

func (t tokenizer) encrypt(token token) string {
	crypt := jwt.New(jwt.SigningMethodHS256)
	crypt.Claims = token.toClaims()
	encrypted, err := crypt.SignedString(t.key)
	if err != nil {
		panic(err)
	}

	return encrypted
}

func (t tokenizer) decrypt(encryptedToken string) token {
	tok, err := jwt.Parse(encryptedToken, jwt.Keyfunc(func(*jwt.Token) (interface{}, error) {
		return t.key, nil
	}))
	if err != nil {
		panic(err)
	}

	return newTokenFromClaims(tok.Claims)
}

func (t tokenizer) validate(tok, expected token) bool {
	if ok := tok.hasAudiences(expected.Audiences); !ok {
		return false
	}

	if ok := tok.hasScopes(expected.Scopes); !ok {
		return false
	}

	return true
}

type token struct {
	UserID    string
	ClientID  string
	Scopes    []string
	Audiences []string
}

func newTokenFromClaims(claims map[string]interface{}) token {
	t := token{}

	if userID, ok := claims["user_id"].(string); ok {
		t.UserID = userID
	}

	if clientID, ok := claims["client_id"].(string); ok {
		t.ClientID = clientID
	}

	if scopes, ok := claims["scope"].([]interface{}); ok {
		var s []string
		for _, scope := range scopes {
			s = append(s, scope.(string))
		}

		t.Scopes = s
	}

	if audiences, ok := claims["aud"].(string); ok {
		t.Audiences = strings.Split(audiences, " ")
	}

	return t
}

func (t token) toClaims() map[string]interface{} {
	claims := make(map[string]interface{})

	if len(t.UserID) > 0 {
		claims["user_id"] = t.UserID
	}

	if len(t.ClientID) > 0 {
		claims["client_id"] = t.ClientID
	}

	claims["scope"] = t.Scopes
	claims["aud"] = strings.Join(t.Audiences, " ")

	return claims
}

func (t token) hasScopes(scopes []string) bool {
	for _, scope := range scopes {
		if !contains(t.Scopes, scope) {
			return false
		}
	}
	return true
}

func (t token) hasAudiences(audiences []string) bool {
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
