package clickup

import (
	"context"
	"strings"

	"github.com/gebv/asap-tools/storage"
	"go.uber.org/zap"
)

func NewStorage(s *storage.Storage) *Storage {
	return &Storage{
		Storage: s,
		log:     zap.L().Named("clickup_storage"),
	}
}

// - method Get<ModelName> - returns model by ID
// - method GetStateOf<ModelName> - returns status of loading of changes for the model
// - method Upsert<ModelName> - create or overwrite model by ID
type Storage struct {
	*storage.Storage
	log *zap.Logger
}

func (s *Storage) LoadSubTasks(ctx context.Context, task *Task) {
	model := (*Task)(nil)
	cname := model.CollectionName()

	iter := s.FirestoreClient().Collection(cname).Where("ParentTaskID", "==", s.DocRef(task)).Documents(ctx)
	res := s.Iterate(iter, model)

	for idx := range res {
		task.SubTasks = append(task.SubTasks, res[idx].(*Task))
	}
}

// a new model instance and call GetModel
func (s *Storage) GetTask(ctx context.Context, modelID string) *Task {
	model := NewWithID(TaskModel, modelID).(*Task)
	s.GetModel(ctx, model)
	return model
}

// alias to UpsertModel
func (s *Storage) UpsertTask(ctx context.Context, model *Task) error {
	return s.UpsertModel(ctx, model)
}

// alias to DeleteModel
func (s *Storage) DeleteTask(ctx context.Context, model *Task) error {
	return s.DeleteModel(ctx, model)
}

func (s *Storage) FetchListTasks(ctx context.Context, list []*DocRef) []*Task {
	res := []*Task{}
	for idx := range list {
		model := &Task{}
		err := s.LoadToModel(ctx, list[idx], model)
		if err != nil {
			s.log.Warn("Failed find task by ID", zap.Error(err), zap.String("task_id", list[idx].ID))
			continue
		}

		res = append(res, model)
	}
	return res
}

// a new model instance and call GetModel
func (s *Storage) GetFolder(ctx context.Context, modelID string) *Folder {
	model := NewWithID(FolderModel, modelID).(*Folder)
	s.GetModel(ctx, model)
	return model
}

// alias to UpsertModel
func (s *Storage) UpsertFolder(ctx context.Context, model *Folder) error {
	return s.UpsertModel(ctx, model)
}

// alias to DeleteModel
func (s *Storage) DeleteFolder(ctx context.Context, model *Folder) error {
	return s.DeleteModel(ctx, model)
}

// a new model instance and call GetModel
func (s *Storage) GetList(ctx context.Context, modelID string) *List {
	model := NewWithID(ListModel, modelID).(*List)
	s.GetModel(ctx, model)
	return model
}

// alias to UpsertModel
func (s *Storage) UpsertList(ctx context.Context, model *List) error {
	return s.UpsertModel(ctx, model)
}

// alias to DeleteModel
func (s *Storage) DeleteList(ctx context.Context, model *List) error {
	return s.DeleteModel(ctx, model)
}

// a new model instance and call GetModel
func (s *Storage) GetTeam(ctx context.Context, modelID string) *Team {
	model := NewWithID(TeamModel, modelID).(*Team)
	s.GetModel(ctx, model)
	return model
}

// alias to UpsertModel
func (s *Storage) UpsertTeam(ctx context.Context, model *Team) error {
	return s.UpsertModel(ctx, model)
}

// alias to DeleteModel
func (s *Storage) DeleteTeam(ctx context.Context, model *Team) error {
	return s.DeleteModel(ctx, model)
}
func (s *Storage) AllTeamTasks(ctx context.Context, teamID string) []*Task {
	cname := TaskModel.CollectionName()

	iter := s.FirestoreClient().Collection(cname).
		Where("TeamRef", "==", s.DocRef(NewWithID(TeamModel, teamID))).Documents(ctx)
	res := s.Iterate(iter, TaskModel)
	list := []*Task{}
	for idx := range res {
		list = append(list, res[idx].(*Task))
	}
	return list
}

// a new model instance and call GetModel
func (s *Storage) GetMember(ctx context.Context, modelID string) *Member {
	model := NewWithID(MemberModel, modelID).(*Member)
	s.GetModel(ctx, model)
	return model
}

// alias to UpsertModel
func (s *Storage) UpsertMember(ctx context.Context, model *Member) error {
	return s.UpsertModel(ctx, model)
}

// alias to DeleteModel
func (s *Storage) DeleteMember(ctx context.Context, model *Member) error {
	return s.DeleteModel(ctx, model)
}
func (s *Storage) FetchListMembers(ctx context.Context, list []*DocRef) []*Member {
	res := []*Member{}
	for idx := range list {
		model := &Member{}
		err := s.LoadToModel(ctx, list[idx], model)
		if err != nil {
			s.log.Warn("Failed find member by ID", zap.Error(err), zap.String("member_id", list[idx].ID))
			continue
		}

		res = append(res, model)
	}
	return res
}
func (s *Storage) MemberByEmail(ctx context.Context, email string) *Member {
	model := &Member{}
	cname := model.CollectionName()
	iter := s.FirestoreClient().Collection(cname).Where("Email", "==", strings.ToLower(email)).Documents(ctx)
	res := s.Iterate(iter, model)
	if len(res) == 0 {
		return model
	}
	if len(res) > 1 {
		s.log.Warn("More than one member with the same email address was found (but return the first one in list)", zap.String("email", email), zap.Int("found_num_members", len(res)))
	}
	return res[0].(*Member)
}
