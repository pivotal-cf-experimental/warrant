package server

import (
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
)

const (
	origin = "uaa"
	schema = "urn:scim:schemas:core:1.0"
)

var schemas = []string{schema}

// Config is the set of configuration values to provide
// to the fake server.
type Config struct {
	PublicKey string
}

// UAA is a fake implementation of the UAA HTTP service.
type UAA struct {
	server    *httptest.Server
	users     *users
	clients   *clients
	groups    *groups
	tokenizer tokenizer

	defaultScopes []string
	publicKey     string
}

// NewUAA returns a new UAA initialized with the given Config.
func NewUAA(config Config) *UAA {
	router := mux.NewRouter()
	server := &UAA{
		server: httptest.NewUnstartedServer(router),
		defaultScopes: []string{
			"scim.read",
			"cloudcontroller.admin",
			"password.write",
			"scim.write",
			"openid",
			"cloud_controller.write",
			"cloud_controller.read",
			"doppler.firehose",
		},
		publicKey: config.PublicKey,
		tokenizer: newTokenizer("this is the encryption key"), // TODO: use a real RSA key
		users:     newUsers(),
		clients:   newClients(),
		groups:    newGroups(),
	}

	router.HandleFunc("/Users", server.createUser).Methods("POST")
	router.HandleFunc("/Users", server.findUsers).Methods("GET")
	router.HandleFunc("/Users/{guid}", server.getUser).Methods("GET")
	router.HandleFunc("/Users/{guid}", server.deleteUser).Methods("DELETE")
	router.HandleFunc("/Users/{guid}", server.updateUser).Methods("PUT")
	router.HandleFunc("/Users/{guid}/password", server.updateUserPassword).Methods("PUT")

	router.HandleFunc("/oauth/clients", server.createClient).Methods("POST")
	router.HandleFunc("/oauth/clients/{guid}", server.getClient).Methods("GET")
	router.HandleFunc("/oauth/clients/{guid}", server.deleteClient).Methods("DELETE")

	router.HandleFunc("/oauth/token", server.oAuthToken).Methods("POST")
	router.HandleFunc("/oauth/authorize", server.oAuthAuthorize).Methods("POST")

	router.HandleFunc("/token_key", server.getTokenKey).Methods("GET")

	router.HandleFunc("/Groups", server.createGroup).Methods("POST")
	router.HandleFunc("/Groups", server.listGroups).Methods("GET")
	router.HandleFunc("/Groups/{guid}", server.getGroup).Methods("GET")
	router.HandleFunc("/Groups/{guid}", server.deleteGroup).Methods("DELETE")

	return server
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
	s.users.clear()
	s.clients.clear()
	s.groups.clear()
}

// URL returns the url that the server is hosted on.
func (s *UAA) URL() string {
	return s.server.URL
}

// SetDefaultScopes allows the default scopes applied to a
// user to be configured.
func (s *UAA) SetDefaultScopes(scopes []string) {
	s.defaultScopes = scopes
} // TODO: move this configuration onto the Config

// ClientTokenFor returns a client token with the given id,
// scopes, and audiences.
func (s *UAA) ClientTokenFor(clientID string, scopes, audiences []string) string {
	// TODO: remove from API so that tokens are fetched like
	// they would be with a real UAA server.

	return s.tokenizer.encrypt(token{
		ClientID:  clientID,
		Scopes:    scopes,
		Audiences: audiences,
	})
}

// UserTokenFor returns a user token with the given id,
// scopes, and audiences.
func (s *UAA) UserTokenFor(userID string, scopes, audiences []string) string {
	// TODO: remove from API so that tokens are fetched like
	// they would be with a real UAA server.

	return s.tokenizer.encrypt(token{
		UserID:    userID,
		Scopes:    scopes,
		Audiences: audiences,
	})
}

func (s *UAA) validateToken(encryptedToken string, audiences, scopes []string) bool {
	t := s.tokenizer.decrypt(encryptedToken)

	return s.tokenizer.validate(t, token{
		Audiences: audiences,
		Scopes:    scopes,
	})
}

func generateID() string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return guid.String()
}
