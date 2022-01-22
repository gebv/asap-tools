package api

import "net/http"

//////////////////////
// List Spaces
//////////////////////

type ListSpacesRequest struct {
	TeamID string
}

func (r *ListSpacesRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/team/" + r.TeamID + "/space"

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type ListSpacesResponse struct {
	responseMetadata
	Spaces []ListSpacesResponse_SpaceItem `json:"spaces"`
}

type ListSpacesResponse_SpaceItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	Statuses []struct {
		Status     string `json:"status"`
		Type       string `json:"type"`
		Orderindex int    `json:"orderindex"`
		Color      string `json:"color"`
	} `json:"statuses"`
	MultipleAssignees bool `json:"multiple_assignees"`
	Features          struct {
		DueDates struct {
			Enabled            bool `json:"enabled"`
			StartDate          bool `json:"start_date"`
			RemapDueDates      bool `json:"remap_due_dates"`
			RemapClosedDueDate bool `json:"remap_closed_due_date"`
		} `json:"due_dates"`
		TimeEstimates struct {
			Enabled bool `json:"enabled"`
		} `json:"time_estimates"`
	} `json:"features"`
}

//////////////////////
// Space By ID
//////////////////////

type SpaceByIDRequest struct {
	SpaceID string
}

func (r *SpaceByIDRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/space/" + r.SpaceID

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type SpaceByIDResponse struct {
	responseMetadata
	ID       string `json:"id"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
	Statuses []struct {
		Status     string `json:"status"`
		Type       string `json:"type"`
		Orderindex int    `json:"orderindex"`
		Color      string `json:"color"`
	} `json:"statuses"`
	MultipleAssignees bool `json:"multiple_assignees"`
	Features          struct {
		DueDates struct {
			Enabled            bool `json:"enabled"`
			StartDate          bool `json:"start_date"`
			RemapDueDates      bool `json:"remap_due_dates"`
			RemapClosedDueDate bool `json:"remap_closed_due_date"`
		} `json:"due_dates"`
		TimeEstimates struct {
			Enabled bool `json:"enabled"`
		} `json:"time_estimates"`
	} `json:"features,omitempty"`
}
