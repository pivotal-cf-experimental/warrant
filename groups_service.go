package warrant

import (
	"encoding/json"
	"net/http"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
	"github.com/pivotal-cf-experimental/warrant/internal/network"
)

type GroupsService struct {
	config Config
}

func NewGroupsService(config Config) GroupsService {
	return GroupsService{
		config: config,
	}
}

func (gs GroupsService) Create(displayName, token string) (Group, error) {
	resp, err := newNetworkClient(gs.config).MakeRequest(network.Request{
		Method:        "POST",
		Path:          "/Groups",
		Authorization: network.NewTokenAuthorization(token),
		Body: network.NewJSONRequestBody(documents.CreateGroupRequest{
			DisplayName: displayName,
			Schemas:     Schemas,
		}),
		AcceptableStatusCodes: []int{http.StatusCreated},
	})
	if err != nil {
		return Group{}, translateError(err)
	}

	var response documents.GroupResponse
	err = json.Unmarshal(resp.Body, &response)
	if err != nil {
		return Group{}, MalformedResponseError{err}
	}

	return newGroupFromResponse(gs.config, response), nil
}

func newGroupFromResponse(config Config, response documents.GroupResponse) Group {
	return Group{
		ID:          response.ID,
		DisplayName: response.DisplayName,
		Version:     response.Meta.Version,
		CreatedAt:   response.Meta.Created,
		UpdatedAt:   response.Meta.LastModified,
	}
}
