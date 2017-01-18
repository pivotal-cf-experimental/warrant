package tokens

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pivotal-cf-experimental/warrant/internal/server/common"
	"github.com/pivotal-cf-experimental/warrant/internal/server/domain"
)

type authorizeHandler struct {
	tokens  *domain.Tokens
	users   *domain.Users
	clients *domain.Clients
	urlFinder  urlFinder
}

func (h authorizeHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Accept") != "application/json" {
		h.redirectToLogin(w)
		return
	}

	requestQuery := req.URL.Query()
	clientID := requestQuery.Get("client_id")
	responseType := requestQuery.Get("response_type")

	if responseType != "token" {
		h.redirectToLogin(w)
		return
	}

	client, ok := h.clients.Get(clientID)
	if !ok {
		common.JSONError(w, http.StatusUnauthorized, fmt.Sprintf("No client with requested id: %s", clientID), "invalid_client")
		return
	}

	req.ParseForm()
	userName := req.Form.Get("username")

	user, ok := h.users.GetByName(userName)
	if !ok {
		common.JSONError(w, http.StatusNotFound, fmt.Sprintf("User %s does not exist", userName), "scim_resource_not_found")
		return
	}

	if req.Form.Get("source") != "credentials" {
		h.redirectToLogin(w)
		return
	}

	if req.Form.Get("password") != user.Password {
		h.redirectToLogin(w)
		return
	}

	scopes := []string{}
	requestedScopes := strings.Split(req.Form.Get("scope"), " ")
	for _, requestedScope := range requestedScopes {
		if contains(h.tokens.DefaultScopes, requestedScope) {
			scopes = append(scopes, requestedScope)
		}
	}

	for _, approvedScope := range scopes {
		if !contains(client.Autoapprove, approvedScope) {
			// TODO: when the scopes requested are not contained
			// in the autoapprove list, the correct behavior is
			// to return a 200 with some JSON explaining to the
			// requestor that the scopes need to be approved by
			// the user. We are obviously not doing this. This
			// needs to be implemented.
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	token := domain.Token{
		UserID:    user.ID,
		ClientID:  clientID,
		Scopes:    scopes,
		Issuer:    fmt.Sprintf("%s/oauth/token", h.urlFinder.URL()),
	}

	updateAudiences(token)

	t := h.tokens.Encrypt(token)

	redirectURI := requestQuery.Get("redirect_uri")

	query := url.Values{
		"token_type":   []string{"bearer"},
		"access_token": []string{t},
		"expires_in":   []string{"599"},
		"scope":        []string{strings.Join(scopes, " ")},
		"jti":          []string{"ad0efc96-ed29-43ef-be75-85a4b4f105b5"},
	}
	location := fmt.Sprintf("%s#%s", redirectURI, query.Encode())

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusFound)
}

func contains(scopeList []string, requestedScope string) bool {
	for _, scope := range scopeList {
		if scope == requestedScope {
			return true
		}
	}

	return false
}

func updateAudiences(token domain.Token) []string {
	audMap:= make(map[string]string)

	for _, scope := range token.Scopes {
		index := strings.Index(scope, ".")

		var audience string
		if index == -1 {
			audience = scope
		} else {
			audience = scope[:index]
		}
		audMap[audience] = audience
	}

	for _, authority := range token.Authorities {
		index := strings.Index(authority, ".")

		var audience string
		if index == -1 {
			audience = authority
		} else {
			audience = authority[:index]
		}
		audMap[audience] = audience
	}

	audiences := make([]string, 0, len(audMap) + 1)

	if token.ClientID != "" {
		audiences = append(audiences, token.ClientID)
	}

	for aud := range audMap {
		audiences = append(audiences, aud)

	}
	return audiences
}

func (h authorizeHandler) redirectToLogin(w http.ResponseWriter) {
	w.Header().Set("Location", "/login")
	w.WriteHeader(http.StatusFound)
}
