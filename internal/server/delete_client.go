package server

import (
	"net/http"
	"regexp"
	"strings"
)

func (s *UAA) deleteClient(w http.ResponseWriter, req *http.Request) {
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	if len(token) == 0 {
		s.Error(w, http.StatusUnauthorized, "Full authentication is required to access this resource", "unauthorized")
		return
	}

	if ok := s.ValidateToken(token, []string{"clients"}, []string{"clients.write"}); !ok {
		s.Error(w, http.StatusForbidden, "Invalid token does not contain resource id (clients)", "access_denied")
		return
	}

	matches := regexp.MustCompile(`/oauth/clients/(.*)$`).FindStringSubmatch(req.URL.Path)
	id := matches[1]

	if ok := s.clients.delete(id); !ok {
		panic("foo")
	}

	w.WriteHeader(http.StatusOK)
}
