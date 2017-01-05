package documents

type Member struct {
	// The alias of the identity provider that authenticated
	// this user. "uaa" is an internal UAA user.
	Origin string `json:"origin"`

	// Type is either "USER" or "GROUP".
	Type string `json:"type"`

	// Value is the globally-unique ID of the member entity,
	// either a user ID or another group ID.
	Value string `json:"value"`
}

type MemberResponse struct {
	// The alias of the identity provider that authenticated
	// this user. "uaa" is an internal UAA user.
	Origin string `json:"origin"`

	// Type is either "USER" or "GROUP".
	Type string `json:"type"`

	// Value is the globally-unique ID of the member entity,
	// either a user ID or another group ID.
	Value string `json:"value"`
}

type MemberListResponse struct {
	// Schemas is the list of schemas for this API request.
	Schemas []string `json:"schemas"`

	// Resources is a list of member resources.
	Resources []MemberResponse `json:"resources"`

	// StartIndex is the index number to start at when returning
	// the list of resources.
	StartIndex int `json:"startIndex"`

	// ItemsPerPage is the number of items to return in the
	// list of resources.
	ItemsPerPage int `json:"itemsPerPage"`

	// TotalResults is the total number of resources that match
	// the list query.
	TotalResults int `json:"totalResults"`
}
