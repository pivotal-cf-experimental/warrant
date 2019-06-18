package testserver

import (
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/pivotal-cf-experimental/warrant/internal/server/clients"
	"github.com/pivotal-cf-experimental/warrant/internal/server/common"
	"github.com/pivotal-cf-experimental/warrant/internal/server/domain"
	"github.com/pivotal-cf-experimental/warrant/internal/server/groups"
	"github.com/pivotal-cf-experimental/warrant/internal/server/tokens"
	"github.com/pivotal-cf-experimental/warrant/internal/server/users"
)

var defaultScopes = []string{
	"scim.read",
	"cloudcontroller.admin",
	"password.write",
	"scim.write",
	"openid",
	"cloud_controller.write",
	"cloud_controller.read",
	"doppler.firehose",
	"notification_preferences.write",
	"notification_preferences.read",
}

// UAA is a fake implementation of the UAA HTTP service.
type UAA struct {
	server  *httptest.Server
	users   *domain.Users
	clients *domain.Clients
	groups  *domain.Groups
	tokens  *domain.Tokens

	publicKey  string
	privateKey string
}

// NewUAA returns a new UAA initialized with the given Config.
func NewUAA() *UAA {
	privateKey := common.TestPrivateKey
	publicKey := common.TestPublicKey

	tokensCollection := domain.NewTokens(publicKey, privateKey, defaultScopes)
	usersCollection := domain.NewUsers()
	clientsCollection := domain.NewClients()
	groupsCollection := domain.NewGroups()

	router := mux.NewRouter()
	uaa := &UAA{
		server:     httptest.NewUnstartedServer(router),
		tokens:     tokensCollection,
		users:      usersCollection,
		clients:    clientsCollection,
		groups:     groupsCollection,
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	tokenRouter := tokens.NewRouter(
		tokensCollection,
		usersCollection,
		clientsCollection,
		publicKey,
		privateKey,
		uaa)

	router.Handle("/Users{a:.*}", users.NewRouter(usersCollection, tokensCollection))
	router.Handle("/Groups{a:.*}", groups.NewRouter(groupsCollection, tokensCollection))
	router.Handle("/oauth/clients{a:.*}", clients.NewRouter(clientsCollection, tokensCollection))
	router.Handle("/oauth{a:.*}", tokenRouter)
	router.Handle("/token_key{a:.*}", tokenRouter)

	return uaa
}

func (s *UAA) PublicKey() string {
	return s.publicKey
}

func (s *UAA) PrivateKey() string {
	return s.privateKey
}

// Start will cause the HTTP server to bind to a port
// and start serving requests.
func (s *UAA) Start() {
	s.server.Start()
}

// Close will cause the HTTP server to stop serving
// requests and close its connection.
func (s *UAA) Close() {
	s.server.Close()
}

// Reset will clear all internal resource state within
// the server. This means that all users, clients, and
// groups will be deleted.
func (s *UAA) Reset() {
	s.users.Clear()
	s.clients.Clear()
	s.groups.Clear()
}

// URL returns the url that the server is hosted on.
func (s *UAA) URL() string {
	return s.server.URL
}

// SetDefaultScopes allows the default scopes applied to a
// user to be configured.
func (s *UAA) SetDefaultScopes(scopes []string) {
	s.tokens.DefaultScopes = scopes
} // TODO: move this configuration onto the Config

// ResetDefaultScopes resets the default scopes back to their
// original values.
func (s *UAA) ResetDefaultScopes() {
	s.tokens.DefaultScopes = defaultScopes
}

// UserTokenFor returns a user token with the given id,
// scopes, and audiences.
func (s *UAA) UserTokenFor(userID string, scopes, audiences []string) string {
	// TODO: remove from API so that tokens are fetched like
	// they would be with a real UAA server.

	return s.tokens.Encrypt(domain.Token{
		UserID:    userID,
		Scopes:    scopes,
		Audiences: audiences,
	})
}
