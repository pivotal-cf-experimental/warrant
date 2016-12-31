package warrant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
	"github.com/pivotal-cf-experimental/warrant/internal/network"
)

// TODO: Pagination for List

// GroupsService provides access to common group actions. Using this service,
// you can create, delete, fetch and list group resources.
type GroupsService struct {
	config Config
}

// NewGroupsService returns a GroupsService initialized with the given Config.
func NewGroupsService(config Config) GroupsService {
	return GroupsService{
		config: config,
	}
}

// Create will make a request to UAA to create a new group resource with the given
// DisplayName. A token with the "scim.write" scope is required.
func (gs GroupsService) Create(displayName, token string) (Group, error) {
	resp, err := newNetworkClient(gs.config).MakeRequest(network.Request{
		Method:        "POST",
		Path:          "/Groups",
		Authorization: network.NewTokenAuthorization(token),
		Body: network.NewJSONRequestBody(documents.CreateGroupRequest{
			DisplayName: displayName,
			Schemas:     schemas,
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

func (gs GroupsService) AddMember(groupID, memberID, token string) error {
	_, err := newNetworkClient(gs.config).MakeRequest(network.Request{
		Method:        "POST",
		Path:          fmt.Sprintf("/Groups/%s/members", groupID),
		Authorization: network.NewTokenAuthorization(token),
		Body: network.NewJSONRequestBody(documents.Member{
			Origin: "uaa",
			Type:   "USER",
			Value:  memberID,
		}),
		AcceptableStatusCodes: []int{http.StatusCreated},
	})
	if err != nil {
		return translateError(err)
	}

	return nil
}

func (gs GroupsService) ListMembers(groupID, token string) ([]Member, error) {
	resp, err := newNetworkClient(gs.config).MakeRequest(network.Request{
		Method:                "GET",
		Path:                  fmt.Sprintf("/Groups/%s/members", groupID),
		Authorization:         network.NewTokenAuthorization(token),
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return nil, translateError(err)
	}

	var response []documents.Member
	err = json.Unmarshal(resp.Body, &response)
	if err != nil {
		return nil, MalformedResponseError{err}
	}

	var memberList []Member
	for _, m := range response {
		memberList = append(memberList, Member{
			ID: m.Value,
		})
	}

	return memberList, nil
}

func (gs GroupsService) RemoveMember(groupID, memberID, token string) error {
	_, err := newNetworkClient(gs.config).MakeRequest(network.Request{
		Method:                "DELETE",
		Path:                  fmt.Sprintf("/Groups/%s/members/%s", groupID, memberID),
		Authorization:         network.NewTokenAuthorization(token),
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return translateError(err)
	}

	return nil
}

// Get will make a request to UAA to fetch the group resource with the matching id.
// A token with the "scim.read" scope is required.
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

// List wil make a request to UAA to list the groups that match the given Query.
// A token with the "scim.read" scope is required.
func (gs GroupsService) List(query Query, token string) ([]Group, error) {
	requestPath := url.URL{
		Path: "/Groups",
		RawQuery: url.Values{
			"filter": []string{query.Filter},
			"sortBy": []string{query.SortBy},
		}.Encode(),
	}

	resp, err := newNetworkClient(gs.config).MakeRequest(network.Request{
		Method:                "GET",
		Path:                  requestPath.String(),
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

// Delete will make a request to UAA to delete the group resource with the matching id.
// A token with the "scim.write" scope is required.
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
