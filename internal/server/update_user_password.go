package server

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

func (s *UAA) updateUserPassword(w http.ResponseWriter, req *http.Request) {
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	matches := regexp.MustCompile(`/Users/(.*)/password$`).FindStringSubmatch(req.URL.Path)
	id := matches[1]

	user, ok := s.users.get(id)
	if !ok {
		s.Error(w, http.StatusUnauthorized, "Not authorized", "access_denied")
		return
	}

	var document documents.ChangePasswordRequest
	err := json.NewDecoder(req.Body).Decode(&document)
	if err != nil {
		panic(err)
	}

	if !s.canUpdateUserPassword(id, token, user.Password, document.OldPassword) {
		s.Error(w, http.StatusUnauthorized, "Not authorized", "access_denied")
		return
	}

	user.Password = document.Password
	s.users.update(user)
}

func (s *UAA) canUpdateUserPassword(userID, tokenHeader, existingPassword, givenPassword string) bool {
	if s.validateToken(tokenHeader, []string{"password"}, []string{"password.write"}) {
		return true
	}

	t := s.tokenizer.decrypt(tokenHeader)
	if t.UserID == userID && existingPassword == givenPassword {
		return true
	}

	return false
}
