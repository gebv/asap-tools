package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

//////////////////////
// Search Task Comments
//////////////////////

type SearchCommentsInTaskRequest struct {
	TaskID      string
	StartTaskID string
	StartTimeTS int64
}

func (r *SearchCommentsInTaskRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/task/" + r.TaskID + "/comment"

	q := reqURL.Query()
	if r.StartTaskID != "" {
		q.Add("start_id", r.StartTaskID)
	}
	if r.StartTimeTS > 0 {
		q.Add("start", fmt.Sprint(r.StartTimeTS))
	}
	reqURL.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type SearchCommentsInTaskResponse struct {
	Comments []struct {
		ID          string `json:"id"`
		CommentText string `json:"comment_text"`
		Assignee    *struct {
			ID             int64  `json:"id"`
			Username       string `json:"username"`
			Email          string `json:"email"`
			Color          string `json:"color"`
			Initials       string `json:"initials"`
			ProfilePicture string `json:"profilePicture"`
		} `json:"assignee"`
		AssignedBy *struct {
			ID             int64  `json:"id"`
			Username       string `json:"username"`
			Email          string `json:"email"`
			Color          string `json:"color"`
			Initials       string `json:"initials"`
			ProfilePicture string `json:"profilePicture"`
		} `json:"assigned_by"`
		User struct {
			ID             int64  `json:"id"`
			Username       string `json:"username"`
			Email          string `json:"email"`
			Color          string `json:"color"`
			Initials       string `json:"initials"`
			ProfilePicture string `json:"profilePicture"`
		} `json:"user"`
		DateAt int64 `json:"date,string"`
	} `json:"comments"`
}

//////////////////////
// Add Comment To Task
//////////////////////

type AddCommentToTaskRequest struct {
	TaskID           string
	CommentText      string
	AssignToMemberID string
}

func (r *AddCommentToTaskRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/task/" + r.TaskID + "/comment"

	dat := map[string]interface{}{
		"comment_text": r.CommentText,
	}

	if r.AssignToMemberID != "" {
		dat["assignee"] = r.AssignToMemberID
	}

	datBytes, _ := json.Marshal(dat)

	req, _ := http.NewRequest(http.MethodPost, reqURL.String(), bytes.NewReader(datBytes))
	return req
}

type AddCommentToTaskResponse struct {
	responseMetadata
}
