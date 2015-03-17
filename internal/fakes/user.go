package fakes

import (
	"errors"
	"time"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

type User struct {
	ID        string
	UserName  string
	CreatedAt time.Time
	UpdatedAt time.Time
	Version   int
	Emails    []string
	Groups    []Group
	Active    bool
	Verified  bool
	Origin    string
	Password  string
}

func newUserFromCreateDocument(request documents.CreateUserRequest) User {
	var emails []string
	for _, email := range request.Emails {
		emails = append(emails, email.Value)
	}

	now := time.Now()
	return User{
		ID:        GenerateID(),
		UserName:  request.UserName,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   0,
		Emails:    emails,
		Groups:    make([]Group, 0),
		Active:    true,
		Verified:  false,
		Origin:    Origin,
	}
}

func newUserFromUpdateDocument(request documents.UpdateUserRequest) User {
	var emails []string
	for _, email := range request.Emails {
		emails = append(emails, email.Value)
	}

	return User{
		ID:        request.ID,
		UserName:  request.UserName,
		CreatedAt: request.Meta.Created,
		UpdatedAt: request.Meta.LastModified,
		Version:   request.Meta.Version,
		Emails:    emails,
		Groups:    make([]Group, 0),
		Active:    true,
		Verified:  false,
		Origin:    Origin,
	}
}

func (u User) ToDocument() documents.UserResponse {
	var emails []documents.Email
	for _, email := range u.Emails {
		emails = append(emails, documents.Email{
			Value: email,
		})
	}

	var groups []documents.Group
	for _ = range u.Groups {
		groups = append(groups, documents.Group{})
	}

	return documents.UserResponse{
		Schemas:  Schemas,
		ID:       u.ID,
		UserName: u.UserName,
		Meta: documents.Meta{
			Version:      u.Version,
			Created:      u.CreatedAt,
			LastModified: u.UpdatedAt,
		},
		Emails:   emails,
		Groups:   groups,
		Active:   u.Active,
		Verified: u.Verified,
		Origin:   u.Origin,
	}
}

func (u User) Validate() (bool, error) {
	if len(u.Emails) == 0 {
		return false, errors.New("An email must be provided.")
	}

	for _, email := range u.Emails {
		if email == "" {
			return false, errors.New("[Assertion failed] - this String argument must have text; it must not be null, empty, or blank")
		}
	}

	return true, nil
}
