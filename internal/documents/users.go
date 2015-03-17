package documents

import "time"

type CreateUserRequest struct {
	UserName string   `json:"userName"`
	Name     UserName `json:"name"`
	Emails   []Email  `json:"emails"`
}

type UpdateUserRequest struct {
	Schemas    []string `json:"schemas"`
	ID         string   `json:"id"`
	UserName   string   `json:"userName"`
	ExternalID string   `json:"externalId"`
	Name       UserName `json:"name"`
	Emails     []Email  `json:"emails"`
	Meta       Meta     `json:"meta"`
}

type UserResponse struct {
	Schemas    []string `json:"schemas"`
	ID         string   `json:"id"`
	ExternalID string   `json:"externalId"`
	UserName   string   `json:"userName"`
	Name       UserName `json:"name"`
	Emails     []Email  `json:"emails"`
	Meta       Meta     `json:"meta"`
	Groups     []Group  `json:"groups"`
	Active     bool     `json:"active"`
	Verified   bool     `json:"verified"`
	Origin     string   `json:"origin"`
}

type UserName struct {
	Formatted  string `json:"formatted"`
	FamilyName string `json:"familyName"`
	GivenName  string `json:"givenName"`
	MiddleName string `json:"middleName"`
}

type Email struct {
	Value string `json:"value"`
}

type Meta struct {
	Version      int       `json:"version"`
	Created      time.Time `json:"created"`
	LastModified time.Time `json:"lastModified"`
}

type Group struct{}
