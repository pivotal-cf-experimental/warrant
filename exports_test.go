package warrant

type TestRequestArguments struct {
	Method                string
	Path                  string
	Token                 string
	IfMatch               string
	Body                  interface{}
	AcceptableStatusCodes []int
}

func NewRequestArguments(args TestRequestArguments) requestArguments {
	return requestArguments{
		Method:  args.Method,
		Path:    args.Path,
		Token:   args.Token,
		IfMatch: args.IfMatch,
		Body:    args.Body,
		AcceptableStatusCodes: args.AcceptableStatusCodes,
	}
}

func (client Client) MakeRequest(requestArgs requestArguments) (int, []byte, error) {
	return client.makeRequest(requestArgs)
}
