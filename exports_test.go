package warrant

import "net/url"

type TestRequestArguments struct {
	Method                string
	Path                  string
	Token                 string
	IfMatch               string
	Body                  requestBody
	AcceptableStatusCodes []int
	DoNotFollowRedirects  bool
}

func NewRequestArguments(args TestRequestArguments) requestArguments {
	return requestArguments{
		Method:  args.Method,
		Path:    args.Path,
		Token:   args.Token,
		IfMatch: args.IfMatch,
		Body:    args.Body,
		AcceptableStatusCodes: args.AcceptableStatusCodes,
		DoNotFollowRedirects:  args.DoNotFollowRedirects,
	}
}

func (client Client) MakeRequest(requestArgs requestArguments) (response, error) {
	return client.makeRequest(requestArgs)
}

func NewJSONRequestBody(body interface{}) jsonRequestBody {
	return jsonRequestBody{body}
}

func NewFormRequestBody(body url.Values) formRequestBody {
	return formRequestBody(body)
}
