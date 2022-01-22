package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"
)

func WebhookVerifier(secret string) func(r *http.Request) bool {
	return func(r *http.Request) bool {
		signature := r.Header.Get("X-Signature")

		hash := hmac.New(sha256.New, []byte(secret))
		io.Copy(hash, r.Body)
		generatedSignature := hex.EncodeToString(hash.Sum(nil))

		return signature == generatedSignature
	}
}

func ParseWebhook(req *http.Request) *WebhookMessage {
	contentType := req.Header.Get("Content-type")
	model := &WebhookMessage{}

	if contentType == "application/json" {
		if err := decodeFromJsonTo(req.Body, model); err != nil {
			return nil
		}
	} else {
		bodyBytes, _ := io.ReadAll(req.Body)
		zap.L().Warn("Received not json webhook", zap.String("uri", req.URL.String()), zap.String("method", req.Method), zap.String("body_raw", string(bodyBytes)), zap.String("content_type", contentType))
		return nil
	}

	return model
}

type WebhookMessage struct {
	WebhookID    string          `json:"webhook_id"`
	EventName    string          `json:"event"`
	HistoryItems json.RawMessage `json:"history_items"`
	TaskID       *string         `json:"task_id"`
	SpaceID      *string         `json:"space_id"`
	ListID       *string         `json:"list_id"`
	FolderID     *string         `json:"folder_id"`
}
