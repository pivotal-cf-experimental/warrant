package warrant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

type Client struct {
	ID                   string
	Scope                []string
	ResourceIDs          []string
	Authorities          []string
	AuthorizedGrantTypes []string
	AccessTokenValidity  time.Duration
}

type ClientsService struct {
	config Config
}

func NewClientsService(config Config) ClientsService {
	return ClientsService{
		config: config,
	}
}

func (cs ClientsService) Create(client Client, secret, token string) error {
	_, err := New(cs.config).makeRequest(requestArguments{
		Method: "POST",
		Path:   "/oauth/clients",
		Token:  token,
		Body:   jsonRequestBody{client.ToDocument(secret)},
		AcceptableStatusCodes: []int{http.StatusCreated},
	})
	if err != nil {
		panic(err)
	}

	return nil
}

func (cs ClientsService) Get(id, token string) (Client, error) {
	resp, err := New(cs.config).makeRequest(requestArguments{
		Method: "GET",
		Path:   fmt.Sprintf("/oauth/clients/%s", id),
		Token:  token,
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return Client{}, err
	}

	var document documents.ClientResponse
	err = json.Unmarshal(resp.Body, &document)
	if err != nil {
		panic(err)
	}

	return newClientFromDocument(document), nil
}

func newClientFromDocument(document documents.ClientResponse) Client {
	return Client{
		ID:                   document.ClientID,
		Scope:                document.Scope,
		ResourceIDs:          document.ResourceIDs,
		Authorities:          document.Authorities,
		AuthorizedGrantTypes: document.AuthorizedGrantTypes,
		AccessTokenValidity:  time.Duration(document.AccessTokenValidity) * time.Second,
	}
}

func (c Client) ToDocument(secret string) documents.CreateClientRequest {
	client := documents.CreateClientRequest{
		ClientID:             c.ID,
		ClientSecret:         secret,
		AccessTokenValidity:  int(c.AccessTokenValidity.Seconds()),
		Scope:                make([]string, 0),
		ResourceIDs:          make([]string, 0),
		Authorities:          make([]string, 0),
		AuthorizedGrantTypes: make([]string, 0),
	}
	client.Scope = append(client.Scope, c.Scope...)
	client.ResourceIDs = append(client.ResourceIDs, c.ResourceIDs...)
	client.Authorities = append(client.Authorities, c.Authorities...)
	client.AuthorizedGrantTypes = append(client.AuthorizedGrantTypes, c.AuthorizedGrantTypes...)

	return client
}
