package models

// APIResponse
// This models contains the response of the bing API.
type APIResponse struct {
	Images []*Image `json:"images"`
}

// Image ...
type Image struct {
	StartDate     string
	FullStartDate string
	EndDate       string
	URL           string
	URLBase       string
	Copyright     string
	CopyrightLink string
	Title         string
}
