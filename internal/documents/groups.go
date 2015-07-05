package documents

type CreateGroupRequest struct {
	Schemas     []string `json:"schemas"`
	DisplayName string   `json:"displayName"`
}

type GroupResponse struct {
	ID          string   `json:"id"`
	Schemas     []string `json:"schemas"`
	DisplayName string   `json:"displayName"`
	Meta        Meta     `json:"meta"`
}

type GroupListResponse struct {
	Resources    []GroupResponse `json:"resources"`
	StartIndex   int             `json:"startIndex"`
	ItemsPerPage int             `json:"itemsPerPage"`
	TotalResults int             `json:"totalResults"`
	Schemas      []string        `json:"schemas"`
}
