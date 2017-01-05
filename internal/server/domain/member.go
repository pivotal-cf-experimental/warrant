package domain

import (
	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

type Member struct {
	Origin string
	Type   string
	Value  string
}

func NewMemberFromDocument(request documents.Member) Member {
	return Member{
		Type:   request.Type,
		Value:  request.Value,
		Origin: request.Origin,
	}
}

func (m Member) ToDocument() documents.MemberResponse {
	return documents.MemberResponse{
		Type:   m.Type,
		Value:  m.Value,
		Origin: m.Origin,
	}
}
