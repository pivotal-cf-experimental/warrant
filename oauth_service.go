package warrant

import (
	"net/http"
	"net/url"
)

type OAuthService struct {
	config Config
}

func NewOAuthService(config Config) OAuthService {
	return OAuthService{
		config: config,
	}
}

func (os OAuthService) GetToken(username, password, client, redirectURI string) (string, error) {
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

	resp, err := NewClient(os.config).makeRequest(req)
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
