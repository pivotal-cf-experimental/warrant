package server

import (
	"encoding/json"
	"net/http"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

type groupsList []group

func (gl groupsList) toDocument() documents.GroupListResponse {
	doc := documents.GroupListResponse{
		ItemsPerPage: 100,
		StartIndex:   1,
		TotalResults: len(gl),
		Schemas:      schemas,
	}

	for _, group := range gl {
		doc.Resources = append(doc.Resources, group.toDocument())
	}

	return doc
}
func (s *UAAServer) listGroups(w http.ResponseWriter, req *http.Request) {
	list := groupsList(s.groups.all())

	response, err := json.Marshal(list.toDocument())
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}
