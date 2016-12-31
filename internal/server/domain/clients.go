package domain

type Clients struct {
	store map[string]Client
}

func NewClients() *Clients {
	return &Clients{
		store: make(map[string]Client),
	}
}

func (collection Clients) All() []Client {
	var clients []Client
	for _, c := range collection.store {
		clients = append(clients, c)
	}
	return clients
}

func (collection Clients) Add(c Client) {
	collection.store[c.ID] = c
}

func (collection Clients) Get(id string) (Client, bool) {
	c, ok := collection.store[id]
	return c, ok
}

func (collection *Clients) Clear() {
	collection.store = make(map[string]Client)
}

func (collection Clients) Delete(id string) bool {
	_, ok := collection.store[id]
	delete(collection.store, id)
	return ok
}

type ByName ClientsList

func (clients ByName) Len() int {
	return len(clients)
}

func (clients ByName) Swap(i, j int) {
	clients[i], clients[j] = clients[j], clients[i]
}

func (clients ByName) Less(i, j int) bool {
	return clients[i].Name < clients[j].Name
}

type ByID ClientsList

func (clients ByID) Len() int {
	return len(clients)
}

func (clients ByID) Swap(i, j int) {
	clients[i], clients[j] = clients[j], clients[i]
}

func (clients ByID) Less(i, j int) bool {
	return clients[i].ID < clients[j].ID
}
