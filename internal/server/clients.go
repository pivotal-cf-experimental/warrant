package server

type clients struct {
	store map[string]client
}

func newClients() *clients {
	return &clients{
		store: make(map[string]client),
	}
}

func (collection clients) add(c client) {
	collection.store[c.ID] = c
}

func (collection clients) get(id string) (client, bool) {
	c, ok := collection.store[id]
	return c, ok
}

func (collection *clients) clear() {
	collection.store = make(map[string]client)
}

func (collection clients) delete(id string) bool {
	_, ok := collection.store[id]
	delete(collection.store, id)
	return ok
}
