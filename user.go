package warrant

import (
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
}

func newUserFromResponse(config Config, response documents.UserResponse) User {
	var emails []string
	for _, email := range response.Emails {
		emails = append(emails, email.Value)
	}

	return User{
		ID:        response.ID,
		UserName:  response.UserName,
		Emails:    emails,
		CreatedAt: response.Meta.Created,
		UpdatedAt: response.Meta.LastModified,
		Active:    response.Active,
		Verified:  response.Verified,
		Origin:    response.Origin,
	}
}
