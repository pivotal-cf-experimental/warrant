package fakes

import (
	"time"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

type Group struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int
}

func newGroupFromCreateDocument(document documents.CreateGroupRequest) Group {
	now := time.Now().UTC()
	return Group{
		ID:          GenerateID(),
		DisplayName: document.DisplayName,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     0,
	}
}

func (g Group) ToDocument() documents.GroupResponse {
	return documents.GroupResponse{
		Schemas:     Schemas,
		ID:          g.ID,
		DisplayName: g.DisplayName,
		Meta: documents.Meta{
			Version:      g.Version,
			Created:      g.CreatedAt,
			LastModified: g.UpdatedAt,
		},
	}
}
