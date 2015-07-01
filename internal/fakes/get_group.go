package fakes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func (s *UAAServer) GetGroup(w http.ResponseWriter, req *http.Request) {
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	if ok := s.ValidateToken(token, []string{"scim"}, []string{"scim.read"}); !ok {
		s.Error(w, http.StatusUnauthorized, "Full authentication is required to access this resource", "unauthorized")
		return
	}

	matches := regexp.MustCompile(`/Groups/(.*)$`).FindStringSubmatch(req.URL.Path)
	id := matches[1]

	user, ok := s.groups.Get(id)
	if !ok {
		s.NotFound(w, fmt.Sprintf("Group %s does not exist", id))
		return
	}

	response, err := json.Marshal(user.ToDocument())
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
