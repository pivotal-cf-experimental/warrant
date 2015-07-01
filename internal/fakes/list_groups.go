package fakes

import (
	"encoding/json"
	"net/http"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

type GroupsList []Group

func (gl GroupsList) ToDocument() documents.GroupListResponse {
	doc := documents.GroupListResponse{
		ItemsPerPage: 100,
		StartIndex:   1,
		TotalResults: len(gl),
		Schemas:      Schemas,
	}

	for _, group := range gl {
		doc.Resources = append(doc.Resources, group.ToDocument())
	}

	return doc
}
func (s *UAAServer) ListGroups(w http.ResponseWriter, req *http.Request) {
	list := GroupsList(s.groups.All())

	response, err := json.Marshal(list.ToDocument())
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
