package fakes

import (
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
)

const (
	Origin = "uaa"
	Schema = "urn:scim:schemas:core:1.0"
)

var Schemas = []string{Schema}

type UAAServer struct {
	server    *httptest.Server
	tokenizer Tokenizer
	users     *Users
	clients   *Clients
}

func NewUAAServer() *UAAServer {
	router := mux.NewRouter()
	server := &UAAServer{
		server:    httptest.NewUnstartedServer(router),
		tokenizer: NewTokenizer("this is the encryption key"),
		users:     NewUsers(),
		clients:   NewClients(),
	}

	router.HandleFunc("/Users", server.CreateUser).Methods("POST")
	router.HandleFunc("/Users/{guid}", server.GetUser).Methods("GET")
	router.HandleFunc("/Users/{guid}", server.DeleteUser).Methods("DELETE")
	router.HandleFunc("/Users/{guid}", server.UpdateUser).Methods("PUT")
	router.HandleFunc("/Users/{guid}/password", server.UpdateUserPassword).Methods("PUT")

	router.HandleFunc("/oauth/authorize", server.OAuthAuthorize).Methods("POST")
	router.HandleFunc("/oauth/clients/{guid}", server.GetClient).Methods("GET")
	router.HandleFunc("/oauth/clients", server.CreateClient).Methods("POST")
	router.HandleFunc("/oauth/token", server.OAuthToken).Methods("POST")

	return server
}

func (s *UAAServer) Start() {
	s.server.Start()
}

func (s *UAAServer) Close() {
	s.server.Close()
}

func (s *UAAServer) Reset() {
	s.users.Clear()
	s.clients.Clear()
}

func (s *UAAServer) URL() string {
	return s.server.URL
}

func (s *UAAServer) ClientTokenFor(clientID string, scopes, audiences []string) string {
	return s.tokenizer.Encrypt(Token{
		ClientID:  clientID,
		Scopes:    scopes,
		Audiences: audiences,
	})
}

func (s *UAAServer) UserTokenFor(userID string, scopes, audiences []string) string {
	return s.tokenizer.Encrypt(Token{
		UserID:    userID,
		Scopes:    scopes,
		Audiences: audiences,
	})
}

func (s *UAAServer) ValidateToken(encryptedToken string, audiences, scopes []string) bool {
	token := s.tokenizer.Decrypt(encryptedToken)

	return s.tokenizer.Validate(token, Token{
		Audiences: audiences,
		Scopes:    scopes,
	})
}

func GenerateID() string {
	guid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return guid.String()
}
