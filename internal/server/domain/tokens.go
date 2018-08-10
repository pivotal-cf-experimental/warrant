package domain

import (
	"crypto/x509"
	"encoding/pem"
	"errors"

	"github.com/dgrijalva/jwt-go"
)

type Tokens struct {
	DefaultScopes []string
	PublicKey     string
	PrivateKey    string
}

func NewTokens(publicKey, privateKey string, defaultScopes []string) *Tokens {
	return &Tokens{
		DefaultScopes: defaultScopes,
		PublicKey:     publicKey,
		PrivateKey:    privateKey,
	}
}

func (t Tokens) Encrypt(token Token) string {
	crypt := jwt.NewWithClaims(jwt.SigningMethodRS256, token.toClaims())
	crypt.Header["kid"] = "legacy-token-key"

	block, _ := pem.Decode([]byte(t.PrivateKey))
	if block == nil {
		panic("failed to decode PEM block containing public key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	encrypted, err := crypt.SignedString(privateKey)
	if err != nil {
		panic(err)
	}

	return encrypted
}

func (t Tokens) Decrypt(encryptedToken string) (Token, error) {
	tok, err := jwt.ParseWithClaims(encryptedToken, jwt.MapClaims{}, jwt.Keyfunc(func(token *jwt.Token) (interface{}, error) {
		switch token.Method {
		case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512:
			block, _ := pem.Decode([]byte(t.PublicKey))
			if block == nil {
				return nil, errors.New("failed to decode PEM block containing public key")
			}

			publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
			if err != nil {
				return nil, err
			}

			return publicKey, nil
		default:
			return nil, errors.New("Unsupported signing method")
		}
	}))
	if err != nil {
		return Token{}, err
	}

	if !tok.Valid {
		return Token{}, errors.New("token is invalid")
	}

	return newTokenFromClaims(tok.Claims.(jwt.MapClaims)), nil
}

func (t Tokens) Validate(encryptedToken string, expectedToken Token) bool {
	decryptedToken, err := t.Decrypt(encryptedToken)
	if err != nil {
		return false
	}

	return t.validate(decryptedToken, expectedToken)
}

func (t Tokens) validate(tok, expected Token) bool {
	if ok := tok.hasAudiences(expected.Audiences); !ok {
		return false
	}

	if ok := tok.hasScopes(expected.Scopes); !ok {
		return false
	}

	if ok := tok.hasAuthorities(expected.Authorities); !ok {
		return false
	}

	return true
}
