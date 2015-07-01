package fakes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

func (s *UAAServer) CreateGroup(w http.ResponseWriter, req *http.Request) {
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	if ok := s.ValidateToken(token, []string{"scim"}, []string{"scim.write"}); !ok {
		s.Error(w, http.StatusUnauthorized, "Full authentication is required to access this resource", "unauthorized")
		return
	}

	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	var document documents.CreateGroupRequest
	err = json.Unmarshal(requestBody, &document)
	if err != nil {
		panic(err)
	}

	if _, ok := s.groups.GetByName(document.DisplayName); ok {
		s.Error(w, http.StatusConflict, fmt.Sprintf("A group with displayName: %s already exists.", document.DisplayName), "scim_resource_already_exists")
		return
	}

	group := newGroupFromCreateDocument(document)
	s.groups.Add(group)

	response, err := json.Marshal(group.ToDocument())
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
