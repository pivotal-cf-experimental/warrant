package warrant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	ClientID string   `json:"client_id"`
	UserID   string   `json:"user_id"`
	Scopes   []string `json:"scope"`
}

type TokensService struct{}

func NewTokensService(config Config) TokensService {
	return TokensService{}
}

func (ts TokensService) Decode(token string) (Token, error) {
	segments := strings.Split(token, ".")
	if len(segments) != 3 {
		return Token{}, InvalidTokenError{fmt.Errorf("invalid number of segments in token (%d/3)", len(segments))}
	}

	claims, err := jwt.DecodeSegment(segments[1])
	if err != nil {
		return Token{}, InvalidTokenError{fmt.Errorf("claims cannot be decoded: %s", err)}
	}

	t := Token{}
	err = json.Unmarshal(claims, &t)
	if err != nil {
		return Token{}, InvalidTokenError{fmt.Errorf("token cannot be parsed: %s", err)}
	}

	return t, nil
}
