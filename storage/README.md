# Storage

The firestore (Google Firebase database) is used as the persistent storage

`Model` (next model) is a `interface` for custom structure with user data. It stores user data and specifies in which collection is stored. It also implements the method of creating an instance of itself.

The model can be standard or with a custom ID.

## Standart Model

Example of a standard model
```go
type Event struct {
    StdModel // This is a required field for standard storage models

    // user fields
    EventName string
    EventAt time.Time
    ...
}

```
Any user storage model must
- method that specifies of model storage collection name `CollectionName() string`
- method of creating a model instance `NewModel() StoreModel`


So we add two methods

```go
var _ Model = (*Event)(nil)
var EventModel = (*Event)(nil)

func (*Event) NewModel() StoreModel {
	return &Event{}
}


func (t *Event) CollectionName() string {
	return "<firestore collection name>"
}
```

Done. The model can be used in Storage. Because it implements the interface `Model`.

```go
ctx := context.TODO()
s := NewStorage(/*TODO*/)

// event := (*Event)(nil).NewModel().(*Event)
// event.SetModelID("123)
// or
// event := EventModel.NewModel().(*Event)
// event.SetModelID("123)
// or
// event := &Event{StdModel: StdModel{PrimaryData: PrimaryData{ID: "123"}}}
// or
// event := NewWithID((*Event)(nil), "123")
// or

event := NewWithID(EventModel, "123")

// get model by primary ID
s.GetModel(ctx, event)

// change
event.EventAt = time.Now().UTC()

// apply changes
s.UpsertModel(ctx, event)

// remove if exists model
if event.Exsits() {
    s.DeleteModel(ctx, event)
}
```

## Custom ID Model

```go
type Event struct {
    ModelCustomID // This is a required field for storage models with custom logic for ID
    ...
}
```

And in addition to the mandatory fields must ba added
- implement `ModelID() string`
- implement `SetModelID(string)`

