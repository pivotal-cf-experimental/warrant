package server

import (
	"time"

	"github.com/pivotal-cf-experimental/warrant/internal/documents"
)

type user struct {
	ID            string
	ExternalID    string
	UserName      string
	FormattedName string
	FamilyName    string
	GivenName     string
	MiddleName    string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Version       int
	Emails        []string
	Groups        []group
	Active        bool
	Verified      bool
	Origin        string
	Password      string
}

func newUserFromCreateDocument(request documents.CreateUserRequest) user {
	var emails []string
	for _, email := range request.Emails {
		emails = append(emails, email.Value)
	}

	now := time.Now().UTC()
	return user{
		ID:        generateID(),
		UserName:  request.UserName,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   0,
		Emails:    emails,
		Groups:    make([]group, 0),
		Active:    true,
		Verified:  false,
		Origin:    origin,
	}
}

func newUserFromUpdateDocument(request documents.UpdateUserRequest) user {
	var emails []string
	for _, email := range request.Emails {
		emails = append(emails, email.Value)
	}

	return user{
		ID:            request.ID,
		ExternalID:    request.ExternalID,
		UserName:      request.UserName,
		FormattedName: request.Name.Formatted,
		FamilyName:    request.Name.FamilyName,
		GivenName:     request.Name.GivenName,
		MiddleName:    request.Name.MiddleName,
		CreatedAt:     request.Meta.Created,
		UpdatedAt:     request.Meta.LastModified,
		Version:       request.Meta.Version,
		Emails:        emails,
		Groups:        make([]group, 0),
		Active:        true,
		Verified:      false,
		Origin:        origin,
	}
}

func (u user) toDocument() documents.UserResponse {
	var emails []documents.Email
	for _, email := range u.Emails {
		emails = append(emails, documents.Email{
			Value: email,
		})
	}

	var groups []documents.Group
	for i := 0; i < len(u.Groups); i++ {
		groups = append(groups, documents.Group{})
	}

	return documents.UserResponse{
		Schemas:    schemas,
		ID:         u.ID,
		ExternalID: u.ExternalID,
		UserName:   u.UserName,
		Name: documents.UserName{
			Formatted:  u.FormattedName,
			FamilyName: u.FamilyName,
			GivenName:  u.GivenName,
			MiddleName: u.MiddleName,
		},
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

func (u user) validate() error {
	if len(u.Emails) == 0 {
		return validationError("An email must be provided.")
	}

	for _, email := range u.Emails {
		if email == "" {
			return validationError("[Assertion failed] - this String argument must have text; it must not be null, empty, or blank")
		}
	}

	return nil
}
