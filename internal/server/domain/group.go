package domain

import (
	"time"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
	"github.com/pivotal-cf-experimental/warrant/internal/server/common"
)

type group struct {
	ID          string
	DisplayName string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int
}

func NewGroupFromCreateDocument(document documents.CreateGroupRequest) group {
	now := time.Now().UTC()
	id, err := common.NewUUID()
	if err != nil {
		panic(err)
	}

	return group{
		ID:          id,
		DisplayName: document.DisplayName,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     0,
	}
}

func (g group) ToDocument() documents.GroupResponse {
	return documents.GroupResponse{
		Schemas:     schemas,
		ID:          g.ID,
		DisplayName: g.DisplayName,
		Meta: documents.Meta{
			Version:      g.Version,
			Created:      g.CreatedAt,
			LastModified: g.UpdatedAt,
		},
	}
}
