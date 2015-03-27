package warrant

import "net/url"

type TestRequestArguments struct {
	Method                string
	Path                  string
	Authorization         authorization
	IfMatch               string
	Body                  requestBody
	AcceptableStatusCodes []int
	DoNotFollowRedirects  bool
}

func NewRequestArguments(args TestRequestArguments) requestArguments {
	return requestArguments{
		Method:        args.Method,
		Path:          args.Path,
		Authorization: args.Authorization,
		IfMatch:       args.IfMatch,
		Body:          args.Body,
		AcceptableStatusCodes: args.AcceptableStatusCodes,
		DoNotFollowRedirects:  args.DoNotFollowRedirects,
	}
}

func (w Warrant) MakeRequest(requestArgs requestArguments) (response, error) {
	return w.makeRequest(requestArgs)
}

func NewJSONRequestBody(body interface{}) jsonRequestBody {
	return jsonRequestBody{body}
}

func NewFormRequestBody(body url.Values) formRequestBody {
	return formRequestBody(body)
}

func NewTokenAuthorization(token string) tokenAuthorization {
	return tokenAuthorization(token)
}

func NewBasicAuthorization(username, password string) basicAuthorization {
	return basicAuthorization{
		Username: username,
		Password: password,
	}
}
