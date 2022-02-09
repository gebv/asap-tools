package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

//////////////////////
// Search Tasks In Team
//////////////////////

type SearchTasksInTeamRequest struct {
	TeamID string

	FolderIDs       []string
	SpaceIDs        []string
	ListIDs         []string
	OrderBy         string
	DateUpdatedGtTs int64
	Page            int
	StatuseNames    []string
	AssignUserIDs   []string
	IncludeClosed   bool
	IncludeSubtasks bool
}

func (r *SearchTasksInTeamRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/team/" + r.TeamID + "/task"

	q := reqURL.Query()
	if r.OrderBy != "" {
		q.Add("order_by", r.OrderBy)
	}
	if r.DateUpdatedGtTs > 0 {
		q.Add("date_updated_gt", fmt.Sprint(r.DateUpdatedGtTs))
	}
	if r.Page > 0 {
		q.Add("page", strconv.Itoa(r.Page))
	}
	if len(r.StatuseNames) > 0 {
		q["statuses[]"] = r.StatuseNames
	}
	if len(r.AssignUserIDs) > 0 {
		q["assignees[]"] = r.AssignUserIDs
	}
	if r.IncludeClosed {
		q.Add("include_closed", "true")
	}
	if r.IncludeSubtasks {
		q.Add("subtasks", "true")
	}
	if len(r.FolderIDs) > 0 {
		q["project_ids[]"] = r.FolderIDs
	}
	if len(r.ListIDs) > 0 {
		q["list_ids[]"] = r.ListIDs
	}
	if len(r.SpaceIDs) > 0 {
		q["space_ids[]"] = r.SpaceIDs
	}
	reqURL.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type SearchTasksInTeamResponse struct {
	responseMetadata
	Tasks []Task `json:"tasks"`
}

type Task struct {
	ID          string  `json:"id"`
	CustomID    *string `json:"custom_id"`
	Name        string  `json:"name"`
	TextContent string  `json:"text_content"`
	Description string  `json:"description"`
	Status      struct {
		Status string `json:"status"`
		Type   string `json:"type"`
	} `json:"status"`
	DateCreatedTs  int64   `json:"date_created,string"`
	DateUpdatedTs  int64   `json:"date_updated,string"`
	DateClosedTs   *int64  `json:"date_closed,string"`
	TimeEstimateMs *int64  `json:"time_estimate"`
	DueDate        *int64  `json:"due_date,string"`
	StartDate      *int64  `json:"start_date,string"`
	Parent         *string `json:"parent"`

	Archived bool `json:"archived"`
	Creator  struct {
		ID             int64  `json:"id"`
		Username       string `json:"username"`
		Color          string `json:"color"`
		Email          string `json:"email"`
		ProfilePicture string `json:"profilePicture"`
	} `json:"creator"`
	Assignees []struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		Color    string `json:"color"`
		Initials string `json:"initials"`
		Email    string `json:"email"`
	} `json:"assignees"`

	Tags []struct {
		Name string `json:"name"`
	} `json:"tags"`

	LinkedTasks []struct {
		TaskID      string `json:"task_id"`
		LinkID      string `json:"link_id"`
		DateCreated int64  `json:"date_created,string"`
		Userid      int64  `json:"userid,string"`
	} `json:"linked_tasks"`
	TeamID          string `json:"team_id"`
	URL             string `json:"url"`
	PermissionLevel string `json:"permission_level"`
	List            struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"list"`
	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	Folder struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"folder"`
	Space struct {
		ID string `json:"id"`
	} `json:"space"`

	Priority *struct {
		ID   int `json:"id,string"`
		Name int `json:"priority"`
	} `json:"priority"`
}

func (r *Task) ListLinkedTaskIDs() []string {
	res := []string{}
	for idx := range r.LinkedTasks {
		res = append(res, r.LinkedTasks[idx].TaskID)
	}
	return res
}

func (r Task) ListTags() []string {
	res := []string{}
	for idx := range r.Tags {
		res = append(res, r.Tags[idx].Name)
	}
	return res
}

//////////////////////
// Update Task
//////////////////////
type UpdateTaskRequest struct {
	TaskID      string
	Name        string
	Description string
	// Priority        int
	StatusName      string
	TimeEstimateMs  int64
	AssigneeAdds    []int64
	AssigneeRemoves []int64
	DueDate         int64
	StartDate       int64
}

func (r *UpdateTaskRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/task/" + r.TaskID

	dat := map[string]interface{}{}
	if r.Name != "" {
		dat["name"] = r.Name
	}

	if r.Description != "" {
		dat["description"] = r.Description
	}

	if r.StatusName != "" {
		dat["status"] = r.StatusName
	}

	if r.TimeEstimateMs > 0 {
		dat["time_estimate"] = r.TimeEstimateMs
	}
	if r.TimeEstimateMs == -1 {
		dat["time_estimate"] = nil
	}

	assignees := map[string]interface{}{}
	if len(r.AssigneeAdds) > 0 {
		assignees["add"] = r.AssigneeAdds
	}
	if len(r.AssigneeRemoves) > 0 {
		assignees["rem"] = r.AssigneeRemoves
	}
	if len(assignees) > 0 {
		dat["assignees"] = assignees
	}

	if r.DueDate > 0 {
		dat["due_date"] = fmt.Sprint(r.DueDate)
	}
	if r.DueDate == -1 {
		dat["due_date"] = nil
	}

	if r.StartDate > 0 {
		dat["start_date"] = fmt.Sprint(r.StartDate)
	}
	if r.DueDate == -1 {
		dat["start_date"] = nil
	}

	datBytes, _ := json.Marshal(dat)

	req, _ := http.NewRequest(http.MethodPut, reqURL.String(), bytes.NewReader(datBytes))
	return req
}

type UpdateTaskResponse struct {
	responseMetadata
	Task
}

//////////////////////
// Create Task
//////////////////////

type CreateTaskRequest struct {
	ListID              string
	Name                string
	StatusName          string
	DescriptionMarkdown string
	Tags                []string
	RefTaskID           string
	AssignIDs           []string
}

func (r *CreateTaskRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/list/" + r.ListID + "/task"

	dat := map[string]interface{}{
		"name":                 r.Name,
		"markdown_description": r.DescriptionMarkdown,
	}
	if len(r.Tags) > 0 {
		dat["tags"] = r.Tags
	}
	if r.RefTaskID != "" {
		dat["links_to"] = r.RefTaskID
	}
	if r.StatusName != "" {
		dat["status"] = r.StatusName
	}
	if len(r.AssignIDs) > 0 {
		dat["assignees"] = r.AssignIDs
	}

	datBytes, _ := json.Marshal(dat)

	req, _ := http.NewRequest(http.MethodPost, reqURL.String(), bytes.NewReader(datBytes))
	return req
}

type CreateTaskResponse struct {
	responseMetadata
	TaskID string `json:"id"`
}

//////////////////////
// Search Tasks
//////////////////////

type SearchTasksRequest struct {
	ListID          string
	OrderBy         string
	DateUpdatedGtTs int64
	Page            int
	StatuseNames    []string
	AssignUserIDs   []string

	IncludeClosed   bool
	IncludeSubtasks bool
}

func (r *SearchTasksRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/list/" + r.ListID + "/task"

	q := reqURL.Query()
	if r.OrderBy != "" {
		q.Add("order_by", r.OrderBy)
	}
	if r.DateUpdatedGtTs > 0 {
		q.Add("date_updated_gt", fmt.Sprint(r.DateUpdatedGtTs))
	}
	if r.Page > 0 {
		q.Add("page", strconv.Itoa(r.Page))
	}
	if len(r.StatuseNames) > 0 {
		q["statuses[]"] = r.StatuseNames
	}
	if len(r.AssignUserIDs) > 0 {
		q["assignees[]"] = r.AssignUserIDs
	}
	if r.IncludeClosed {
		q.Add("include_closed", "true")
	}
	if r.IncludeSubtasks {
		q.Add("subtasks", "true")
	}
	reqURL.RawQuery = q.Encode()

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type SearchTasksResponse struct {
	Tasks []Task `json:"tasks"`
}

//////////////////////
// Find task by ID
//////////////////////

type TaskByIDRequest struct {
	TaskID string
}

func (r *TaskByIDRequest) buildRequest() *http.Request {
	reqURL := clickupBaseURL()
	reqURL.Path += "/task/" + r.TaskID

	req, _ := http.NewRequest(http.MethodGet, reqURL.String(), nil)
	return req
}

type TaskByIDResponse struct {
	responseMetadata
	Task
}
