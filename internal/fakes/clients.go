package fakes

import "github.com/pivotal-cf-experimental/warrant/internal/documents"

type Clients struct {
	store map[string]Client
}

func NewClients() *Clients {
	return &Clients{
		store: make(map[string]Client),
	}
}

func (c *Clients) Add(client Client) {
	c.store[client.ID] = client
}

func (c *Clients) Get(id string) (Client, bool) {
	client, ok := c.store[id]
	return client, ok
}

type Client struct {
	ID                   string
	Secret               string
	Scope                []string
	ResourceIDs          []string
	Authorities          []string
	AuthorizedGrantTypes []string
	AccessTokenValidity  int
}

func (c Client) ToDocument() documents.ClientResponse {
	return documents.ClientResponse{
		ClientID:             c.ID,
		Scope:                c.Scope,
		ResourceIDs:          c.ResourceIDs,
		Authorities:          c.Authorities,
		AuthorizedGrantTypes: c.AuthorizedGrantTypes,
		AccessTokenValidity:  c.AccessTokenValidity,
	}
}

func newClientFromDocument(document documents.CreateClientRequest) Client {
	return Client{
		ID:                   document.ClientID,
		Secret:               document.ClientSecret,
		Scope:                document.Scope,
		ResourceIDs:          document.ResourceIDs,
		Authorities:          document.Authorities,
		AuthorizedGrantTypes: document.AuthorizedGrantTypes,
		AccessTokenValidity:  document.AccessTokenValidity,
	}
}
