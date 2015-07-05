package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

func (s *UAAServer) oAuthToken(w http.ResponseWriter, req *http.Request) {
	// TODO: actually check the basic auth values
	_, _, ok := req.BasicAuth()
	if !ok {
		s.Error(w, http.StatusUnauthorized, "An Authentication object was not found in the SecurityContext", "unauthorized")
		return
	}

	err := req.ParseForm()
	if err != nil {
		panic(err)
	}
	clientID := req.Form.Get("client_id")

	scopes := []string{"scim.write", "scim.read", "password.write"}
	t := s.tokenizer.encrypt(token{
		ClientID:  clientID,
		Scopes:    scopes,
		Audiences: []string{"scim", "password"},
	})

	response, err := json.Marshal(documents.TokenResponse{
		AccessToken: t,
		TokenType:   "bearer",
		ExpiresIn:   5000,
		Scope:       strings.Join(scopes, " "),
		JTI:         generateID(),
	})

	w.Write(response)
}