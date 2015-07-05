package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

func (s *UAAServer) createUser(w http.ResponseWriter, req *http.Request) {
	token := strings.TrimPrefix(req.Header.Get("Authorization"), "Bearer ")
	if ok := s.ValidateToken(token, []string{"scim"}, []string{"scim.write"}); !ok {
		s.Error(w, http.StatusUnauthorized, "Full authentication is required to access this resource", "unauthorized")
		return
	}

	requestBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		if contentType == "" {
			contentType = http.DetectContentType(requestBody)
		}
		s.Error(w, http.StatusBadRequest, fmt.Sprintf("Content type '%s' not supported", contentType), "scim")
		return
	}

	var document documents.CreateUserRequest
	err = json.Unmarshal(requestBody, &document)
	if err != nil {
		panic(err)
	}

	if _, ok := s.users.getByName(document.UserName); ok {
		s.Error(w, http.StatusConflict, fmt.Sprintf("Username already in use: %s", document.UserName), "scim_resource_already_exists")
		return
	}

	user := newUserFromCreateDocument(document)
	if err := user.validate(); err != nil {
		s.Error(w, http.StatusBadRequest, err.Error(), "invalid_scim_resource")
		return
	}
	s.users.add(user)

	response, err := json.Marshal(user.toDocument())
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
