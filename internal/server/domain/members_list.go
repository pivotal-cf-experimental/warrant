package domain

import "github.com/pivotal-cf-experimental/warrant/internal/documents"

type MembersList []Member

func (ml MembersList) ToDocument() documents.MemberListResponse {
	doc := documents.MemberListResponse{
		ItemsPerPage: 100,
		StartIndex:   1,
		TotalResults: len(ml),
		Schemas:      schemas,
	}

	for _, member := range ml {
		doc.Resources = append(doc.Resources, member.ToDocument())
	}

	return doc
}
