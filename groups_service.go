package warrant

import (
	"encoding/json"
	"fmt"
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

func (gs GroupsService) Get(id, token string) (Group, error) {
	resp, err := newNetworkClient(gs.config).MakeRequest(network.Request{
		Method:                "GET",
		Path:                  fmt.Sprintf("/Groups/%s", id),
		Authorization:         network.NewTokenAuthorization(token),
		AcceptableStatusCodes: []int{http.StatusOK},
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

func (gs GroupsService) List(token string) ([]Group, error) {
	resp, err := newNetworkClient(gs.config).MakeRequest(network.Request{
		Method:                "GET",
		Path:                  "/Groups",
		Authorization:         network.NewTokenAuthorization(token),
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return []Group{}, translateError(err)
	}

	var response documents.GroupListResponse
	err = json.Unmarshal(resp.Body, &response)
	if err != nil {
		return []Group{}, MalformedResponseError{err}
	}

	var groupList []Group
	for _, groupResponse := range response.Resources {
		groupList = append(groupList, newGroupFromResponse(gs.config, groupResponse))
	}

	return groupList, err
}

func (gs GroupsService) Delete(id, token string) error {
	_, err := newNetworkClient(gs.config).MakeRequest(network.Request{
		Method:                "DELETE",
		Path:                  fmt.Sprintf("/Groups/%s", id),
		Authorization:         network.NewTokenAuthorization(token),
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return translateError(err)
	}

	return nil
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
