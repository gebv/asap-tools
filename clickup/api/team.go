package api

import (
	"net/http"
)

type ListTeamsRequest struct {
}

func (r *ListTeamsRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/team"

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type ListTeamsResponse struct {
	responseMetadata
	Teams []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"teams"`
}
