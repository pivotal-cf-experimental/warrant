package warrant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

const (
	Schema = "urn:scim:schemas:core:1.0"
)

var Schemas = []string{Schema}

type UsersService struct {
	config Config
}

func NewUsersService(config Config) UsersService {
	return UsersService{
		config: config,
	}
}

func (us UsersService) Create(username, email, token string) (User, error) {
	resp, err := New(us.config).makeRequest(requestArguments{
		Method: "POST",
		Path:   "/Users",
		Token:  token,
		Body: jsonRequestBody{documents.CreateUserRequest{
			UserName: username,
			Emails: []documents.Email{
				{Value: email},
			},
		}},
		AcceptableStatusCodes: []int{http.StatusCreated},
	})
	if err != nil {
		return User{}, err
	}

	var response documents.UserResponse
	err = json.Unmarshal(resp.Body, &response)
	if err != nil {
		panic(err)
	}

	return newUserFromResponse(us.config, response), nil
}

func (us UsersService) Get(id, token string) (User, error) {
	resp, err := New(us.config).makeRequest(requestArguments{
		Method: "GET",
		Path:   fmt.Sprintf("/Users/%s", id),
		Token:  token,
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return User{}, err
	}

	var response documents.UserResponse
	err = json.Unmarshal(resp.Body, &response)
	if err != nil {
		panic(err)
	}

	return newUserFromResponse(us.config, response), nil
}

func (us UsersService) Delete(id, token string) error {
	_, err := New(us.config).makeRequest(requestArguments{
		Method: "DELETE",
		Path:   fmt.Sprintf("/Users/%s", id),
		Token:  token,
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return err
	}

	return nil
}

func (us UsersService) Update(user User, token string) (User, error) {
	resp, err := New(us.config).makeRequest(requestArguments{
		Method:  "PUT",
		Path:    fmt.Sprintf("/Users/%s", user.ID),
		Token:   token,
		IfMatch: strconv.Itoa(user.Version),
		Body:    jsonRequestBody{newUpdateUserDocumentFromUser(user)},
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return User{}, err
	}

	var response documents.UserResponse
	err = json.Unmarshal(resp.Body, &response)
	if err != nil {
		panic(err)
	}

	return newUserFromResponse(us.config, response), nil
}

func (us UsersService) SetPassword(id, password, token string) error {
	_, err := New(us.config).makeRequest(requestArguments{
		Method: "PUT",
		Path:   fmt.Sprintf("/Users/%s/password", id),
		Token:  token,
		Body: jsonRequestBody{documents.SetPasswordRequest{
			Password: password,
		}},
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return err
	}

	return nil
}

func (us UsersService) ChangePassword(id, oldPassword, password, token string) error {
	_, err := New(us.config).makeRequest(requestArguments{
		Method: "PUT",
		Path:   fmt.Sprintf("/Users/%s/password", id),
		Token:  token,
		Body: jsonRequestBody{documents.ChangePasswordRequest{
			OldPassword: oldPassword,
			Password:    password,
		}},
		AcceptableStatusCodes: []int{http.StatusOK},
	})
	if err != nil {
		return err
	}

	return nil
}

func (us UsersService) GetToken(username, password, client, redirectURI string) (string, error) {
	query := url.Values{
		"client_id":     []string{"cf"},
		"redirect_uri":  []string{redirectURI},
		"response_type": []string{"token"},
	}

	requestPath := url.URL{
		Path:     "/oauth/authorize",
		RawQuery: query.Encode(),
	}
	req := requestArguments{
		Method: "POST",
		Path:   requestPath.String(),
		Body: formRequestBody{
			"username": []string{username},
			"password": []string{password},
			"source":   []string{"credentials"},
		},
		AcceptableStatusCodes: []int{http.StatusFound},
		DoNotFollowRedirects:  true,
	}

	resp, err := New(us.config).makeRequest(req)
	if err != nil {
		return "", err
	}

	locationURL, err := url.Parse(resp.Headers.Get("Location"))
	if err != nil {
		return "", err
	}

	locationQuery, err := url.ParseQuery(locationURL.Fragment)
	if err != nil {
		return "", err
	}

	return locationQuery.Get("access_token"), nil
}

func newUpdateUserDocumentFromUser(user User) documents.UpdateUserRequest {
	var emails []documents.Email
	for _, email := range user.Emails {
		emails = append(emails, documents.Email{
			Value: email,
		})
	}

	return documents.UpdateUserRequest{
		Schemas:  Schemas,
		ID:       user.ID,
		UserName: user.UserName,
		//ExternalID: user.ExternalID, // TODO
		Name: documents.UserName{
		//Formatted:  user.FormattedName, // TODO
		//FamilyName: user.FamilyName,
		//GivenName:  user.GivenName,
		//MiddleName: user.MiddleName,
		},
		Emails: emails,
		Meta: documents.Meta{
			Version:      user.Version,
			Created:      user.CreatedAt,
			LastModified: user.UpdatedAt,
		},
	}
}
