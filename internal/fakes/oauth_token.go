package fakes

import (
	"encoding/json"
	"net/http"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

func (s *UAAServer) OAuthToken(w http.ResponseWriter, req *http.Request) {
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

	token := s.ClientTokenFor(clientID, []string{}, []string{})

	response, err := json.Marshal(documents.TokenResponse{
		AccessToken: token,
		TokenType:   "bearer",
		ExpiresIn:   5000,
		Scope:       "",
		JTI:         GenerateID(),
	})

	w.Write(response)
}
