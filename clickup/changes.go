package clickup

import (
	"context"
	"fmt"
	"strings"
)

// alias to UpsertModel
func (s *Storage) UpsertLoadStatusOfChangedTeamTasks(ctx context.Context, model *LoadStatusOfChangedTeamTasks) error {
	return s.UpsertModel(ctx, model)
}

// a new model instance and call GetModel
func (s *Storage) GetStateOfLoadChangesForTeamTasks(ctx context.Context, teamID string) *LoadStatusOfChangedTeamTasks {
	model := &LoadStatusOfChangedTeamTasks{TeamID: teamID}
	s.GetModel(ctx, model)
	return model
}

type LoadStatusOfChangedTeamTasks struct {
	StoreModelCustomID
	TeamID  string `firestore:"-"`
	TeamRef *DocRef
	// stores the value from the API
	LastTaskUpdatedAt int64
}

func (m *LoadStatusOfChangedTeamTasks) NewModel() StoreModel {
	return &LoadStatusOfChangedTeamTasks{}
}

func (m *LoadStatusOfChangedTeamTasks) SetModelID(in string) {
	const prefix = "team:"
	if strings.HasPrefix(in, prefix) {
		m.TeamID = in[len(prefix):]
	} else {
		panic(fmt.Sprintf("Invalid format ID %q for %T", in, m))
	}
}

func (t *LoadStatusOfChangedTeamTasks) ModelID() string {
	return fmt.Sprintf("team:%s", t.TeamID)
}

func (LoadStatusOfChangedTeamTasks) CollectionName() string {
	return "clickup_load_status_of_changes"
}

var (
	LoadStatusOfChangedTeamTasksModel            = (*LoadStatusOfChangedTeamTasks)(nil)
	_                                 StoreModel = (*LoadStatusOfChangedTeamTasks)(nil)
)
