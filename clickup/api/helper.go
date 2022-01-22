package api

import (
	"encoding/json"
	"io"
	"net/http"
)

type ListStatuses []Status

type Status struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	// Orderindex int    `json:"orderindex"`
	Color string `json:"color"`
	Type  string `json:"type"`
}

type ResponseMetadata interface {
	StatusOK() bool
	NotFound() bool
	IsStatus(in int) bool

	DecodeOK() bool
}

func decodeFromJsonTo(dat io.Reader, model interface{}) error {
	return json.NewDecoder(dat).Decode(model)
}

type requestBuilder interface {
	buildRequest() *http.Request
}

type setterResponseMetadata interface {
	SetResponseStatusCode(in int)
	SetDecodeErr(err error)
}

var _ ResponseMetadata = (*responseMetadata)(nil)

type responseMetadata struct {
	responseStatusCode int
	decodeErr          error
}

func (r *responseMetadata) SetResponseStatusCode(in int) {
	r.responseStatusCode = in
}

func (r *responseMetadata) SetDecodeErr(err error) {
	r.decodeErr = err
}

func (r responseMetadata) StatusOK() bool {
	return r.responseStatusCode == http.StatusOK
}

func (r responseMetadata) NotFound() bool {
	return r.responseStatusCode == http.StatusNotFound
}

func (r responseMetadata) IsStatus(in int) bool {
	return r.responseStatusCode == in
}

func (r responseMetadata) DecodeOK() bool {
	return r.decodeErr == nil
}
