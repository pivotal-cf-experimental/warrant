package documents

// TokenResponse represents the JSON transport data structure
// for a request that returns a token value.
type TokenResponse struct {
	// AccessToken is the token string used to authenticate
	// with UAA-based services.
	AccessToken string `json:"access_token"`

	// TokenType describes the type of token returned.
	// This value is always "Bearer".
	TokenType string `json:"token_type"`

	// ExpiresIn is the number of seconds until this token
	// expires.
	ExpiresIn int `json:"expires_in"`

	// Scope is a comma separated list of permission values
	// for this token.
	Scope string `json:"scope"`

	// JTI is the unique identifier for this JWT token.
	JTI string `json:"jti"`
}

type TokenKeyResponse struct {
	Alg   string `json:"alg"`
	Value string `json:"value"`
	Kty   string `json:"kty"`
	Use   string `json:"use"`
	N     string `json:"n"`
	E     string `json:"e"`
}
