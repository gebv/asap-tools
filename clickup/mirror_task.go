package clickup

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

var (
	MirrorTaskModel            = (*MirrorTask)(nil)
	_               StoreModel = (*MirrorTask)(nil)
)

// ModelMirrorTaskFor returns prepared MirrorTask model with the corrects populate of the primary information.
// Because the user logic for preparing the storage model ID
func (s *Storage) ModelMirrorTaskFor(taskID, mirrorTaskID string) *MirrorTask {
	return &MirrorTask{
		TaskID:        taskID,
		MirrorTaskID:  mirrorTaskID,
		TaskRef:       s.DocRef(NewWithID(TaskModel, taskID)),
		MirrorTaskRef: s.DocRef(NewWithID(TaskModel, mirrorTaskID)),
	}
}

// alias to GetModel for custom model
func (s *Storage) GetMirrorTask(ctx context.Context, modelID string) *MirrorTask {
	taskID, mirrorTaskID, err := MirrorTaskModel.ParseID(modelID)
	if err != nil {
		s.log.Warn("invalid MirrorTask ID", zap.Error(err), zap.String("model_id", modelID))
		return &MirrorTask{}
	}
	model := s.ModelMirrorTaskFor(taskID, mirrorTaskID)
	s.GetModel(ctx, model)
	return model
}

// alias to UpsertModel
func (s *Storage) UpsertMirrorTask(ctx context.Context, model *MirrorTask) error {
	return s.UpsertModel(ctx, model)
}

// alias to DeleteModel for custom model
func (s *Storage) UnlinkMirroredTask(ctx context.Context, taskID, mirrorTaskID string) error {
	model := s.ModelMirrorTaskFor(taskID, mirrorTaskID)
	return s.DeleteModel(ctx, model)
}

// AllMatchesForMirrorTasks returns union list mirror tasks by task ID and by mirror task ID.
func (s *Storage) AllMatchesForMirrorTasks(ctx context.Context, taskID string) (_ []*MirrorTask, _ bool) {
	cname := MirrorTaskModel.CollectionName()

	list := []*MirrorTask{}
	{
		iter := s.FirestoreClient().Collection(cname).
			Where("MirrorTaskRef", "==", s.DocRef(NewWithID(TaskModel, taskID))).Documents(ctx)
		res := s.Iterate(iter, MirrorTaskModel)
		for idx := range res {
			list = append(list, res[idx].(*MirrorTask))
		}
	}
	len1 := len(list)
	{
		iter := s.FirestoreClient().Collection(cname).
			Where("TaskRef", "==", s.DocRef(NewWithID(TaskModel, taskID))).Documents(ctx)
		res := s.Iterate(iter, MirrorTaskModel)

		for idx := range res {
			list = append(list, res[idx].(*MirrorTask))
		}
	}
	len2 := len(list)

	crossedSync := len1 > 0 && len2 > len1

	return list, crossedSync
}

type MirrorTask struct {
	StoreModelCustomID
	TaskID, MirrorTaskID string `firestore:"-"`
	TaskRef              *DocRef
	Task                 *Task `firestore:"-"`
	MirrorTaskRef        *DocRef
	MirrorTask           *Task `firestore:"-"`
}

func (t *MirrorTask) GetTask(ctx context.Context) *Task {
	if t.Task != nil {
		return t.Task
	}
	t.Task = t.Task.NewModel().(*Task)
	getRefDoc(ctx, t.TaskRef, t.Task)
	return t.Task
}

func (t *MirrorTask) GetMirrorTask(ctx context.Context) *Task {
	if t.MirrorTask != nil {
		return t.MirrorTask
	}
	t.MirrorTask = t.Task.NewModel().(*Task)
	getRefDoc(ctx, t.MirrorTaskRef, t.MirrorTask)
	return t.MirrorTask
}

func (*MirrorTask) NewModel() StoreModel {
	return &MirrorTask{}
}

func (t *MirrorTask) CollectionName() string {
	return "clickup_mirror_tasks"
}

func (t *MirrorTask) ParseID(in string) (taskID, mirrorTaskID string, err error) {
	args := strings.Split(in, ":")
	if len(args) != 4 {
		return "", "", fmt.Errorf("invalid format ID %q", in)
	}
	valid := args[0] == "src" && args[1] != "" && args[2] == "dst" && args[3] != ""
	if !valid {
		return "", "", fmt.Errorf("invalid format ID %q", in)
	}
	return args[1], args[3], nil
}

func (t *MirrorTask) SetModelID(in string) {
	var err error
	t.TaskID, t.MirrorTaskID, err = t.ParseID(in)
	if err != nil {
		panic(err)
	}
}

func (t *MirrorTask) ModelID() string {
	return fmt.Sprintf("src:%s:dst:%s", t.TaskID, t.MirrorTaskID)
}

func (t *Task) MirrorTaskName(ctx context.Context) string {
	priority := ""
	if t.PriorityID != nil {
		priority = fmt.Sprint("!", *t.PriorityID)
	}
	customID := ""
	if t.CustomID != nil {
		customID = "/" + *t.CustomID
	}

	// <ListName>(<CustomID>): (<Priority>)<TaskName>
	return t.GetList(ctx).Name + customID + ": " + priority + " " + t.Name
}

func (t *Task) MirrorTaskDescription() string {
	return `Mirror task from ` + t.URL + `
NOTE: description does not auto-update. Need keep description up to date manually.
* * *
` + t.Description
}
