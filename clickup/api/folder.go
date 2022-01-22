package api

import "net/http"

//////////////////////
// List Folders
//////////////////////

type ListFoldersRequest struct {
	SpaceID string
}

func (r *ListFoldersRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/space/" + r.SpaceID + "/folder"

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type ListFoldersResponse struct {
	responseMetadata
	Folders []struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Orderindex int    `json:"orderindex"`
		Hidden     bool   `json:"hidden"`
		Space      struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"space"`
		TaskCount       string                               `json:"task_count"`
		Archived        bool                                 `json:"archived"`
		Statuses        []ListStatuses                       `json:"statuses"`
		Lists           []ListFoldersResponse_FolderListItem `json:"lists"`
		PermissionLevel string                               `json:"permission_level"`
	} `json:"folders"`
}

type ListFoldersResponse_FolderItem struct {
}

type ListFoldersResponse_FolderListItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TaskCount int64  `json:"task_count"`
	Space     struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Access bool   `json:"access"`
	} `json:"space"`
	Archived        bool           `json:"archived"`
	Statuses        []ListStatuses `json:"statuses"`
	PermissionLevel string         `json:"permission_level"`
}

//////////////////////
// Folder By ID
//////////////////////

type FolderByIDRequest struct {
	FolderID string
}

func (r *FolderByIDRequest) buildRequest() *http.Request {
	// https://api.clickup.com/api/v2/folder/folder_id
	reqURL := clickupBaseURL()
	reqURL.Path += "/folder/" + r.FolderID

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type FolderByIDResponse struct {
	responseMetadata
	ID               string `json:"id"`
	Name             string `json:"name"`
	OverrideStatuses bool   `json:"override_statuses"`
	Hidden           bool   `json:"hidden"`
	Space            struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Access bool   `json:"access"`
	} `json:"space"`
	TaskCount       int64                         `json:"task_count,string"`
	Archived        bool                          `json:"archived"`
	Statuses        []ListStatuses                `json:"statuses"`
	Lists           []FolderByIDResponse_ListItem `json:"lists"`
	PermissionLevel string                        `json:"permission_level"`
}

type FolderByIDResponse_ListItem struct {
	ID               string         `json:"id"`
	Name             string         `json:"name"`
	TaskCount        int64          `json:"task_count"`
	Archived         bool           `json:"archived"`
	OverrideStatuses bool           `json:"override_statuses"`
	Statuses         []ListStatuses `json:"statuses"`
	PermissionLevel  string         `json:"permission_level"`
}
