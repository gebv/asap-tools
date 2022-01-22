package storage

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"go.uber.org/zap"
)

func NewStorage(db *firestore.Client) *Storage {
	return &Storage{
		db:  db,
		log: zap.L().Named("clickup_storage"),
	}
}

type Storage struct {
	db  *firestore.Client
	log *zap.Logger
}

// GetModel looks up model by ID in the store and populate to model.
func (s *Storage) GetModel(ctx context.Context, model Model) error {
	docRef := DocRef(s.db, model)
	err := LoadDocAndPopulate(ctx, docRef, model)
	if err != nil && err != ErrNotFound {
		s.log.Warn("failed find model by ID", withModel(model, zap.Error(err))...)
	}
	return err
}

// UpsertModel updates or creates a model in the store.
// WARN: Model.Exsits() returns false after successfully upsert
func (s *Storage) UpsertModel(ctx context.Context, model Model) error {

	_, err := upsertModel(ctx, s.db.Collection(model.CollectionName()), model)
	if firestoreDocIsAlreadyExists(err) {
		return ErrAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("failed upsert new %T: %w", model, err)
	}
	return nil
}

// DeleteModel deletes the model.
func (s *Storage) DeleteModel(ctx context.Context, model Model) error {
	if model.ModelID() == "" {
		return ErrInvalidModelEmptyID
	}

	_, err := s.db.Collection(model.CollectionName()).Doc(model.ModelID()).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed delete mdoel %T by ID %q in collection %q: %w", model, model.ModelID(),
			model.CollectionName(), err)
	}
	return err
}

var (
	ErrNotFound                = errors.New("not found")
	ErrAlreadyExists           = errors.New("already exists")
	ErrInvalidModelEmptyID     = errors.New("invalid model: empty ID")
	ErrStorageModelMustBeEmpty = errors.New("storage: model must be empty")
)
