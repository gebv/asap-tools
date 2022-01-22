package api

import (
	"fmt"
	"net/http"
)

//////////////////////
// Task Members
//////////////////////

type TaskMembersRequest struct {
	TaskID string
}

func (r *TaskMembersRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/task/" + r.TaskID + "/member"

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type TaskMembersResponse struct {
	Members []Member `json:"members"`
}

type TaskMembersResponse_Member struct {
	Member
}

//////////////////////
// List Members
//////////////////////

type ListMembersRequest struct {
	ListID string
}

func (r *ListMembersRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/list/" + r.ListID + "/member"

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type ListMembersResponse struct {
	responseMetadata
	Members []Member `json:"members"`
}

type Member struct {
	ID             int64  `json:"id"`
	Username       string `json:"username"`
	Email          string `json:"email"`
	Color          string `json:"color"`
	Initials       string `json:"initials"`
	ProfilePicture string `json:"profilePicture"`
}

func (m *Member) IDString() string {
	return fmt.Sprint(m.ID)
}
