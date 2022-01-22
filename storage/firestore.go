package storage

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Iterate iterate through the records and to populate in model.
func (s *Storage) Iterate(iter *firestore.DocumentIterator, factory Model) []Model {
	res, err := IterateAllDocsAndStop(iter, factory)
	if err != nil {
		s.log.Warn("failed iter", withModel(factory.NewModel(), zap.Error(err))...)
		return nil
	}
	return res
}

// LoadToModel looks up firestore document by ID in the store and uses to set fields in model.
func (s *Storage) LoadToModel(ctx context.Context, docRef *firestore.DocumentRef, model Model) error {
	return LoadDocAndPopulate(ctx, docRef, model)
}

func (s *Storage) FirestoreClient() *firestore.Client {
	return s.db
}

func (s *Storage) DocRef(model Model) *DocumentRef {
	return DocRef(s.db, model)
}

// DocRef returns the firestore.DocumentRef based on model.
func DocRef(db *firestore.Client, model Model) *firestore.DocumentRef {
	return db.Collection(model.CollectionName()).Doc(model.ModelID())
}

// IterateAllDocsAndStop iterate through the records and to populate in model.
func IterateAllDocsAndStop(iter *firestore.DocumentIterator, kind Model) ([]Model, error) {
	defer iter.Stop()

	res := []Model{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed iter next item: %w", err)
		}

		model := kind.NewModel()

		if err := PopulateModelFrom(model, doc); err != nil {
			return nil, err
		}
		res = append(res, model)
	}
	return res, nil
}

// LoadDocAndPopulate looks up firestore document by ID in the store and uses to set fields in model.
func LoadDocAndPopulate(ctx context.Context, docRef *firestore.DocumentRef, model Model) error {
	if docRef.ID == "" {
		return ErrInvalidModelEmptyID
	}
	doc, err := docRef.Get(ctx)

	if firestoreDocIsNotFound(err) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed find %T by ID %q in collection %q: %w", model, docRef.ID,
			model.CollectionName(), err)
	}

	if err := PopulateModelFrom(model, doc); err != nil {
		return err
	}

	return nil
}

// PopulateModelFrom set fields in model from firestore document.
func PopulateModelFrom(model Model, doc *firestore.DocumentSnapshot) error {
	if err := doc.DataTo(model); err != nil {
		return fmt.Errorf("failed decode firestore document to model %T: %w", model, err)
	}
	model.SetModelID(doc.Ref.ID)
	model.setDocumentSnapshot(doc)
	return nil
}

// tips: convert go type to firestore type
// - bool converts to Bool.
// - string converts to String.
// - int, int8, int16, int32 and int64 convert to Integer.
// - uint8, uint16 and uint32 convert to Integer. uint, uint64 and uintptr are disallowed,
//   because they may be able to represent values that cannot be represented in an int64,
//   which is the underlying type of a Integer.
// - float32 and float64 convert to Double.
// - []byte converts to Bytes.
// - time.Time and *ts.Timestamp convert to Timestamp. ts is the package
//   "github.com/golang/protobuf/ptypes/timestamp".
// - *latlng.LatLng converts to GeoPoint. latlng is the package
//   "google.golang.org/genproto/googleapis/type/latlng". You should always use
//   a pointer to a LatLng.
// - Slices convert to Array.
// - *firestore.DocumentRef converts to Reference.
// - Maps and structs convert to Map.
// - nils of any type convert to Null.
//
// tips: tag options
// - omitempty: Do not encode this field if it is empty. A value is empty
//   if it is a zero value, or an array, slice or map of length zero.
// - serverTimestamp: The field must be of type time.Time. serverTimestamp
//   is a sentinel token that tells Firestore to substitute the server time
//   into that field. When writing, if the field has the zero value, the
//   server will populate the stored document with the time that the request
//   is processed. However, if the field value is non-zero it won't be saved.
func upsertModel(ctx context.Context, c *firestore.CollectionRef, model Model) (*firestore.WriteResult, error) {
	var doc *firestore.DocumentRef

	customID := model.ModelID()
	if customID != "" {
		doc = c.Doc(customID)
	} else {
		doc = c.NewDoc()
	}

	model.SetModelID(doc.ID)
	return doc.Set(ctx, model)
}

// UpsertIfNotExists helper method for to perform update if the model does not exist in the storage.
// WARN: Model.Exsits() returns false after successfully upsert
func (s *Storage) UpsertIfNotExists(ctx context.Context, model Model) error {
	err := s.GetModel(ctx, model)

	if err == nil {
		// skip for existing model
		return nil
	}

	if err == ErrNotFound {
		return s.UpsertModel(ctx, model)
	}

	s.log.Warn("failed upsert model by ID", withModel(model, zap.Error(err))...)
	return err
}

func firestoreDocIsNotFound(err error) bool {
	return status.Code(err) == codes.NotFound
}

func firestoreDocIsAlreadyExists(err error) bool {
	return status.Code(err) == codes.AlreadyExists
}

type Timestamp = timestamp.Timestamp
type DocumentRef = firestore.DocumentRef
