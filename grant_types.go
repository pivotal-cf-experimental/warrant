package warrant

import "net/url"

type GrantTypes struct {
	ClientCredentials ClientCredentials
	AuthorizationCode AuthorizationCode
	Password          Password
}

type ClientCredentials struct {
	ID     string
	Secret string
}

type Password struct {
	Username string
	Password string
}

type AuthorizationCode struct {
	Code        string
	RedirectURI string
}

func (gt GrantTypes) GenerateForm() url.Values {
	values := url.Values{}
	values.Add("response_type", "token")
	values.Add("grant_type", "client_credentials")
	values.Add("client_id", gt.ClientCredentials.ID)
	values.Add("client_secret", gt.ClientCredentials.Secret)
	return values
}
