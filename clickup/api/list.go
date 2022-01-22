package api

import "net/http"

//////////////////////
// Folder Lists
//////////////////////

type FolderListsRequest struct {
	FolderID string
}

func (r *FolderListsRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/folder/" + r.FolderID + "/list"

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type FolderListsReponse struct {
	responseMetadata
	Lists []FolderListsReponse_ListItem `json:"lists"`
}

type FolderListsReponse_ListItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Folder struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Hidden bool   `json:"hidden"`
		Access bool   `json:"access"`
	} `json:"folder"`
	Space struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Access bool   `json:"access"`
	} `json:"space"`
	Archived bool `json:"archived"`
}

//////////////////////
// Space Folderless Lists
//////////////////////

type SpaceFolderlessListsRequest struct {
	SpaceID string
}

func (r *SpaceFolderlessListsRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/space/" + r.SpaceID + "/list"

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type SpaceFolderlessListsResponse struct {
	responseMetadata
	Lists []SpaceFolderlessListsResponse_ListItem `json:"lists"`
}

type SpaceFolderlessListsResponse_ListItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Folder struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Hidden bool   `json:"hidden"`
		Access bool   `json:"access"`
	} `json:"folder"`
	Space struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Access bool   `json:"access"`
	} `json:"space"`
	Archived bool `json:"archived"`
}

//////////////////////
// List By ID
//////////////////////

type ListByIDRequest struct {
	ListID string
}

func (r *ListByIDRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/list/" + r.ListID

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type ListByIDResponse struct {
	responseMetadata
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Folder  struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Hidden bool   `json:"hidden"`
		Access bool   `json:"access"`
	} `json:"folder"`
	Space struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Access bool   `json:"access"`
	} `json:"space"`
	InboundAddress   string         `json:"inbound_address"`
	Archived         bool           `json:"archived"`
	OverrideStatuses bool           `json:"override_statuses"`
	Statuses         []ListStatuses `json:"statuses"`
}
