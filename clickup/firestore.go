package clickup

import (
	"context"
	"time"

	"github.com/gebv/asap-tools/storage"
	"go.uber.org/zap"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type Timestamp = storage.Timestamp
type DocRef = storage.DocumentRef
type StoreModel = storage.Model
type StdStoreModel = storage.StdModel
type StoreModelCustomID = storage.ModelCustomID

var NewWithID = storage.NewWithID

// Returns Timestamp from timestamp with millesecond.
//
// Clickup uses timestamp with milleseconds.
// https://clickup20.docs.apiary.io/#introduction/faq
// > How are dates formatted in ClickUp?
// ClickUp will always display dates in Unix time in milliseconds. You can use a website like Epoch Converter to convert dates between Unix and human readable date formats.
func TimestampFromTimestampWithMilliseconds(in *int64) *Timestamp {
	if in == nil {
		return nil
	}
	return timestamppb.New(time.Unix(*in/1000, 0))
}

func TimestampNow() *Timestamp {
	return timestamppb.Now()
}

func getRefDoc(ctx context.Context, doc *DocRef, model StoreModel) error {
	err := storage.LoadDocAndPopulate(ctx, doc, model)
	if err != nil {
		zap.L().Warn("failed load doc ref", zap.Error(err))
	}
	return err
}
