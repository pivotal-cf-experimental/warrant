package fakes

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func (s *UAAServer) DeleteGroup(w http.ResponseWriter, req *http.Request) {
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	if ok := s.ValidateToken(token, []string{"scim"}, []string{"scim.write"}); !ok {
		s.Error(w, http.StatusUnauthorized, "Full authentication is required to access this resource", "unauthorized")
		return
	}

	matches := regexp.MustCompile(`/Groups/(.*)$`).FindStringSubmatch(req.URL.Path)
	id := matches[1]

	if ok := s.groups.Delete(id); !ok {
		s.Error(w, http.StatusNotFound, fmt.Sprintf("Group %s does not exist", id), "scim_resource_not_found")
		return
	}

	w.WriteHeader(http.StatusOK)
}
