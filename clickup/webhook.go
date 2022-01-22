package clickup

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/gebv/asap-tools/clickup/api"
)

type WebhookManager struct {
	store         *Storage
	api           *api.API
	webhookSecret string
}

func (s *WebhookManager) Handle(w http.ResponseWriter, req *http.Request) error {
	var body bytes.Buffer
	body.ReadFrom(req.Body)
	req.Body = ioutil.NopCloser(&body)
	if !api.WebhookVerifier(s.webhookSecret)(req) {
		return ErrSignatureMismatch
	}
	msg := api.ParseWebhook(req)
	if msg != nil {
		return ErrInvalidContent
	}

	switch msg.EventName {
	case "taskCreated":
	case "taskUpdated":
	case "taskDeleted":
	case "taskPriorityUpdated":
	case "taskStatusUpdated":
	case "taskAssigneeUpdated":
	case "taskDueDateUpdated":
	case "taskMoved":
	case "taskCommentPosted":
	case "taskCommentUpdated":
	case "taskTimeEstimateUpdated":

	case "listCreated":
	case "listUpdated":
	case "listDeleted":

	case "folderCreated":
	case "folderUpdated":
	case "folderDeleted":

	case "spaceCreated":
	case "spaceUpdated":
	case "spaceDeleted":
	}

	return nil
}

var ErrSignatureMismatch = errors.New("webhook signature mismatch")
var ErrInvalidContent = errors.New("invalid content")
