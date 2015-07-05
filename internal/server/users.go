package server

type users struct {
	store map[string]user
}

func newUsers() *users {
	return &users{
		store: make(map[string]user),
	}
}

func (collection users) add(u user) {
	collection.store[u.ID] = u
}

func (collection users) update(u user) {
	collection.store[u.ID] = u
}

func (collection users) get(id string) (user, bool) {
	u, ok := collection.store[id]
	return u, ok
}

func (collection users) getByName(name string) (user, bool) {
	for _, u := range collection.store {
		if u.UserName == name {
			return u, true
		}
	}

	return user{}, false
}

func (collection users) delete(id string) bool {
	_, ok := collection.store[id]
	delete(collection.store, id)
	return ok
}

func (collection *users) clear() {
	collection.store = make(map[string]user)
}
