package warrant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
	"github.com/pivotal-cf-experimental/warrant/internal/network"
)

// ClientsService provides access to the common client actions. Using this service, you can
// create, delete, or fetch a client. You can also fetch a client token.
type ClientsService struct {
	config Config
}

// NewClientsService returns a ClientsService initialized with the given Config.
func NewClientsService(config Config) ClientsService {
	return ClientsService{
		config: config,
	}
}

// Create will make a request to UAA to register a client with the given client resource and
// secret. A token with the "clients.write" scope is required.
func (cs ClientsService) Create(client Client, secret, token string) error {
	_, err := newNetworkClient(cs.config).MakeRequest(network.Request{
		Method:        "POST",
		Path:          "/oauth/clients",
		Authorization: network.NewTokenAuthorization(token),
		Body:          network.NewJSONRequestBody(client.toDocument(secret)),
		AcceptableStatusCodes: []int{http.StatusCreated},
	})
	if err != nil {
		return translateError(err)
	}

	return nil
}

// Get will make a request to UAA to fetch the client matching the given id.
// A token with the "clients.read" scope is required.
func (cs ClientsService) Get(id, token string) (Client, error) {
	resp, err := newNetworkClient(cs.config).MakeRequest(network.Request{
		Method:                "GET",
		Path:                  fmt.Sprintf("/oauth/clients/%s", id),
		Authorization:         network.NewTokenAuthorization(token),
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return Client{}, translateError(err)
	}

	var document documents.ClientResponse
	err = json.Unmarshal(resp.Body, &document)
	if err != nil {
		return Client{}, MalformedResponseError{err}
	}

	return newClientFromDocument(document), nil
}

// Delete will make a request to UAA to delete the client matching the given id.
// A token with the "clients.write" scope is required.
func (cs ClientsService) Delete(id, token string) error {
	_, err := newNetworkClient(cs.config).MakeRequest(network.Request{
		Method:                "DELETE",
		Path:                  fmt.Sprintf("/oauth/clients/%s", id),
		Authorization:         network.NewTokenAuthorization(token),
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return translateError(err)
	}

	return nil
}

// GetToken will make a request to UAA to retrieve a client token using the
// "client_credentials" grant type. A client id and secret are required.
func (cs ClientsService) GetToken(id, secret string) (string, error) {
	resp, err := newNetworkClient(cs.config).MakeRequest(network.Request{
		Method:        "POST",
		Path:          "/oauth/token",
		Authorization: network.NewBasicAuthorization(id, secret),
		Body: network.NewFormRequestBody(url.Values{
			"client_id":  []string{id},
			"grant_type": []string{"client_credentials"},
		}),
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return "", translateError(err)
	}

	var response documents.TokenResponse
	err = json.Unmarshal(resp.Body, &response)
	if err != nil {
		return "", MalformedResponseError{err}
	}

	return response.AccessToken, nil
}
