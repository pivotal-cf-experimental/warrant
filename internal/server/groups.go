package server

type groups struct {
	store map[string]group
}

func newGroups() *groups {
	return &groups{
		store: make(map[string]group),
	}
}

func (collection groups) add(g group) {
	collection.store[g.ID] = g
}

func (collection groups) update(g group) {
	collection.store[g.ID] = g
}

func (collection groups) get(id string) (group, bool) {
	g, ok := collection.store[id]
	return g, ok
}

func (collection groups) all() []group {
	var groups []group
	for _, g := range collection.store {
		groups = append(groups, g)
	}

	return groups
}

func (collection groups) delete(id string) bool {
	_, ok := collection.store[id]
	delete(collection.store, id)
	return ok
}

func (collection *groups) clear() {
	collection.store = make(map[string]group)
}

func (collection groups) getByName(name string) (group, bool) {
	for _, g := range collection.store {
		if g.DisplayName == name {
			return g, true
		}
	}

	return group{}, false
}
