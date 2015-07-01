package fakes

type Groups struct {
	store map[string]Group
}

func NewGroups() *Groups {
	return &Groups{
		store: make(map[string]Group),
	}
}

func (g Groups) Add(group Group) {
	g.store[group.ID] = group
}

func (g Groups) Update(group Group) {
	g.store[group.ID] = group
}

func (g Groups) Get(id string) (Group, bool) {
	group, ok := g.store[id]
	return group, ok
}

func (g Groups) Delete(id string) bool {
	_, ok := g.store[id]
	delete(g.store, id)
	return ok
}

func (g *Groups) Clear() {
	g.store = make(map[string]Group)
}

func (g Groups) GetByName(name string) (Group, bool) {
	for _, group := range g.store {
		if group.DisplayName == name {
			return group, true
		}
	}

	return Group{}, false
}
