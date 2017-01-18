package tokens

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pivotal-cf-experimental/warrant/internal/server/common"
	"github.com/pivotal-cf-experimental/warrant/internal/server/domain"
)

type urlFinder interface {
	URL() string
}

type tokenHandler struct {
	clients    *domain.Clients
	users      *domain.Users
	urlFinder  urlFinder
	privateKey string
}

func (h tokenHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// TODO: actually check the basic auth values
	clientID, _, ok := req.BasicAuth()
	if !ok {
		common.JSONError(w, http.StatusUnauthorized, "An Authentication object was not found in the SecurityContext", "unauthorized")
		return
	}

	client, ok := h.clients.Get(clientID)
	if !ok {
		common.JSONError(w, http.StatusUnauthorized, fmt.Sprintf("No client with requested id: %s", clientID), "invalid_client")
		return
	}

	err := req.ParseForm()
	if err != nil {
		panic(err)
	}

	var t domain.Token
	t.Scopes = []string{}

	if req.Form.Get("grant_type") == "client_credentials" {
		t.Authorities = client.Authorities
		// This isn't correct - but it will put the resourceID
		// in the audience
		for _, resource := range client.ResourceIDs {
			if resource == "none" {
				continue
			}
			t.Audiences = append(t.Scopes, resource)
		}
	} else {
		user, ok := h.users.GetByName(req.Form.Get("username"))
		if !ok {
			common.JSONError(w, http.StatusNotFound, fmt.Sprintf("User %s does not exist", req.Form.Get("username")), "scim_resource_not_found")
			return
		}

		t.UserID = user.ID
	}

	for _, scope := range client.Scope {
		t.Scopes = append(t.Scopes, scope)
	}

	t.ClientID = clientID
	t.Issuer = fmt.Sprintf("%s/oauth/token", h.urlFinder.URL())
	t.Audiences = updateAudiences(t)

	response, err := json.Marshal(t.ToDocument(h.privateKey))
	if err != nil {
		panic(err)
	}

	w.Write(response)
}
