package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

func clickupBaseURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   "api.clickup.com",
		Path:   "api/v2",
	}
}

type httpClientLogger struct {
	*zap.Logger
}

func (l *httpClientLogger) Printf(msg string, args ...interface{}) {
	l.Debug(fmt.Sprintf(msg, args...))
}

func NewAPI(accessToken string) *API {
	l := zap.L().Named("api")

	httpClient := retryablehttp.NewClient()
	httpClient.Logger = &httpClientLogger{l.Named("http")}

	return &API{
		token:  accessToken,
		client: httpClient.StandardClient(),
		log:    l,
	}
}

type API struct {
	token  string
	client *http.Client
	log    *zap.Logger
}

var _ ResponseMetadata = (*CreateTaskResponse)(nil)
var _ ResponseMetadata = (*UpdateTaskResponse)(nil)
var _ ResponseMetadata = (*AddCommentToTaskResponse)(nil)
var _ ResponseMetadata = (*TaskByIDResponse)(nil)
var _ ResponseMetadata = (*ListTeamsResponse)(nil)
var _ ResponseMetadata = (*ListSpacesResponse)(nil)
var _ ResponseMetadata = (*ListFoldersResponse)(nil)
var _ ResponseMetadata = (*SpaceFolderlessListsResponse)(nil)
var _ ResponseMetadata = (*ListByIDResponse)(nil)
var _ ResponseMetadata = (*FolderByIDResponse)(nil)
var _ ResponseMetadata = (*SpaceByIDResponse)(nil)
var _ ResponseMetadata = (*ListMembersResponse)(nil)
var _ ResponseMetadata = (*SearchTasksInTeamResponse)(nil)
var _ ResponseMetadata = (*ListTeamsResponse)(nil)

func (a *API) CreateTask(ctx context.Context, newTask *CreateTaskRequest) *CreateTaskResponse {
	res := &CreateTaskResponse{}
	a.doRequest(ctx, newTask, res)
	return res
}

func (a *API) UpdateTask(ctx context.Context, updTask *UpdateTaskRequest) *UpdateTaskResponse {
	res := &UpdateTaskResponse{}
	a.doRequest(ctx, updTask, res)
	return res
}

func (a *API) AddCommentToTask(ctx context.Context, newComment *AddCommentToTaskRequest) *AddCommentToTaskResponse {
	res := &AddCommentToTaskResponse{}
	a.doRequest(ctx, newComment, res)
	return res
}

func (a *API) TaskByID(ctx context.Context, taskID string) *TaskByIDResponse {
	req := &TaskByIDRequest{TaskID: taskID}
	res := &TaskByIDResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) ListTeams(ctx context.Context) *ListTeamsResponse {
	req := &ListTeamsRequest{}
	res := &ListTeamsResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) ListSpaces(ctx context.Context, teamID string) *ListSpacesResponse {
	req := &ListSpacesRequest{TeamID: teamID}
	res := &ListSpacesResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) ListFolders(ctx context.Context, spaceID string) *ListFoldersResponse {
	req := &ListFoldersRequest{SpaceID: spaceID}
	res := &ListFoldersResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) SpaceFolderlessLists(ctx context.Context, spaceID string) *SpaceFolderlessListsResponse {
	req := &SpaceFolderlessListsRequest{SpaceID: spaceID}
	res := &SpaceFolderlessListsResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) FolderLists(ctx context.Context, folderID string) *FolderListsReponse {
	req := &FolderListsRequest{FolderID: folderID}
	res := &FolderListsReponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) ListByID(ctx context.Context, listID string) *ListByIDResponse {
	req := &ListByIDRequest{ListID: listID}
	res := &ListByIDResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) FolderByID(ctx context.Context, folderID string) *FolderByIDResponse {
	req := &FolderByIDRequest{FolderID: folderID}
	res := &FolderByIDResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) SpaceByID(ctx context.Context, spaceID string) *SpaceByIDResponse {
	req := &SpaceByIDRequest{SpaceID: spaceID}
	res := &SpaceByIDResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) ListMembersOfList(ctx context.Context, listID string) *ListMembersResponse {
	req := &ListMembersRequest{ListID: listID}
	res := &ListMembersResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) SearchTasksInTeam(ctx context.Context, req *SearchTasksInTeamRequest) *SearchTasksInTeamResponse {
	res := &SearchTasksInTeamResponse{}
	a.doRequest(ctx, req, res)
	return res
}

func (a *API) SearchCommentsInTask(ctx context.Context, taskID string, startTaskID string, startTaskTs int64) *ListTeamsResponse {
	req := &SearchCommentsInTaskRequest{
		TaskID:      taskID,
		StartTaskID: startTaskID,
		StartTimeTS: startTaskTs,
	}
	res := &ListTeamsResponse{}
	a.doRequest(ctx, req, res)
	return res
}

// returns Body which can be read many times and not be closed.
func (a *API) doRequest(ctx context.Context, reqFactory requestBuilder, model setterResponseMetadata) (*http.Response, error) {
	req := reqFactory.buildRequest()

	req.Header.Set("Authorization", a.token)
	req.Header.Set("Content-Type", "application/json")

	req = req.WithContext(ctx)

	res, err := a.client.Do(req)
	a.log.Debug("API request", zap.Error(err), zap.String("status", res.Status), zap.String("uri", req.URL.String()), zap.String("method", req.Method))
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body := &bytes.Buffer{}
	body.ReadFrom(res.Body)
	res.Body = ioutil.NopCloser(body)

	model.SetResponseStatusCode(res.StatusCode)

	// NOTE: Errors responses will return a non-200 status code and a json err message and error code.
	if res.StatusCode != 200 {
		a.log.Debug("Unsuccessful response", zap.String("status", res.Status), zap.String("uri", req.URL.String()), zap.String("method", req.Method), zap.String("body_raw", string(body.String())))
		return res, errors.New("got not-200 status code")
	}

	contentType := req.Header.Get("Content-type")

	if contentType == "application/json" {
		if err := decodeFromJsonTo(res.Body, model); err != nil {
			a.log.Warn(fmt.Sprintf("Failed deocode json to model %T", model), zap.String("uri", req.URL.String()), zap.String("method", req.Method), zap.String("body_raw", string(body.String())), zap.String("content_type", contentType),
				zap.Error(err),
			)

			model.SetDecodeErr(err)

			return res, err
		}
	} else {
		a.log.Warn("Received not json", zap.String("uri", req.URL.String()), zap.String("method", req.Method), zap.String("body_raw", string(body.String())), zap.String("content_type", contentType))

		model.SetDecodeErr(fmt.Errorf("not suppodted content type %q", contentType))
	}

	return res, nil
}
