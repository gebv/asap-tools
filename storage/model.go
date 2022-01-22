package storage

import (
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"go.uber.org/zap"
)

// Model for custom structure with user data. It stores user data and specifies in which collection is stored. It also implements the method of creating an instance of itself.
type Model interface {
	// returns new empy model
	NewModel() Model

	// Exists returns true if model (after retrieving from the database) existing in database.
	Exists() bool
	ModelID() string
	SetModelID(in string)
	// CollectionName returns name of the collection (in firestore) where stored model.
	CollectionName() string
	// ModelUpdatedAt returns update datetime in the database.
	// if model is exists returns valid value
	// if the model does not exist returns empty time.Time (time.Time{}.IsZero() == true)
	ModelUpdatedAt() time.Time
	// ModelUpdatedAt returns create datetime in the database.
	// if model is exists returns valid value
	// if the model does not exist returns empty time.Time (time.Time{}.IsZero() == true)
	ModelCreatedAt() time.Time

	setDocumentSnapshot(in *firestore.DocumentSnapshot)
}

// A helper interface specifying the requirements for custom storage models.
type ModelImpl interface {
	// CollectionName returns name of the collection (in firestore) where stored model.
	CollectionName() string
	NewModel() Model
}

// A helper interface for custom storage models with a custom implementation of model ID storage.
type CustomModelID interface {
	ModelID() string
	SetModelID(in string)
}

// helper method
func NewWithID(kind Model, id string) Model {
	if _, ok := kind.(isStdStorageModel); !ok {
		panic(fmt.Errorf("method clickup.NewWithID supports only for clickup.StdModel but got %T", kind))
	}
	model := kind.NewModel()
	model.SetModelID(id)
	return model
}

type ModelCustomID struct {
	ServiceData
}

type isStdStorageModel interface {
	isStdStorageModel()
}

type StdModel struct {
	ServiceData
	PrimaryData
}

func (*StdModel) isStdStorageModel() {}

type PrimaryData struct {
	ID string `firestore:"-"`
}

func (m *PrimaryData) ModelID() string {
	return m.ID
}

func (m *PrimaryData) SetModelID(in string) {
	if m == nil {
		return
	}
	m.ID = in
}

// Implements a mandatory interface for any storage model.
type ServiceData struct {
	Doc *firestore.DocumentSnapshot `firestore:"-"`
}

func (m *ServiceData) ModelUpdatedAt() time.Time {
	if m == nil {
		return time.Time{}
	}

	if m.Doc == nil {
		return time.Time{}
	}
	return m.Doc.UpdateTime
}

func (m *ServiceData) ModelCreatedAt() time.Time {
	if m == nil {
		return time.Time{}
	}

	if m.Doc == nil {
		return time.Time{}
	}
	return m.Doc.CreateTime
}

func (m *ServiceData) Exists() bool {
	if m == nil {
		return false
	}

	if m.Doc == nil {
		return false
	}
	return m.Doc.Exists()
}

func (m *ServiceData) setDocumentSnapshot(in *firestore.DocumentSnapshot) {
	if m == nil {
		return
	}
	m.Doc = in
}

func withModel(m Model, fields ...zap.Field) []zap.Field {
	return append(fields,
		zap.String("firestore_collection_name", m.CollectionName()),
		zap.String("model_type", fmt.Sprintf("%T", m)),
		zap.String("model_id", m.ModelID()),
	)
}
