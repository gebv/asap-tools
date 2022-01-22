package clickup

import (
	"context"
	"strings"
	"sync"
)

var (
	_ StoreModel = (*Task)(nil)
	_ StoreModel = (*List)(nil)
	_ StoreModel = (*Folder)(nil)
	_ StoreModel = (*Team)(nil)
	_ StoreModel = (*Member)(nil)

	TaskModel   = (*Task)(nil)
	ListModel   = (*List)(nil)
	FolderModel = (*Folder)(nil)
	TeamModel   = (*Team)(nil)
	MemberModel = (*Member)(nil)
)

type Task struct {
	StdStoreModel
	Name          string
	CustomID      *string
	Description   string
	TextContent   string
	StatusType    string
	StatusName    string
	DateCreatedAt *Timestamp
	DateUpdatedAt *Timestamp

	DateClosedAt     *Timestamp
	DueDateAt        *Timestamp
	StartDateAt      *Timestamp
	TimeEstimateMs   *int64
	Archived         bool
	Deleted          bool
	CreatorMemberRef *DocRef
	AssigneesRef     []*DocRef
	URL              string
	ParentTaskRef    *DocRef
	PriorityID       *int
	Tags             []string
	LinkedTasksRef   []*DocRef

	TeamRef   *DocRef
	ListRef   *DocRef
	FolderRef *DocRef

	CreatorMember *Member `firestore:"-"`
	ParentTask    *Task   `firestore:"-"`
	Team          *Team   `firestore:"-"`
	List          *List   `firestore:"-"`
	Folder        *Folder `firestore:"-"`

	// NOTE: concurrent  not safe
	Assignees             []*Member `firestore:"-"`
	lazyLoadAssignees     func()    `firestore:"-" json:"-"`
	lazyLoadAssigneesOnce sync.Once `firestore:"-"`

	SubTasks             []*Task   `firestore:"-"`
	lazyLoadSubTasks     func()    `firestore:"-" json:"-"`
	lazyLoadSubTasksOnce sync.Once `firestore:"-"`

	LinkedTasks             []*Task   `firestore:"-"`
	lazyLoadLinkedTasks     func()    `firestore:"-" json:"-"`
	lazyLoadLinkedTasksOnce sync.Once `firestore:"-"`
}

func (t *Task) TeamID() string {
	return t.TeamRef.ID
}

func (t *Task) EqualUpdatedAt(in *Timestamp) bool {
	return t.DateUpdatedAt.AsTime().Equal(in.AsTime())
}

func (t *Task) IsDeletedOrHidden() bool {
	return t.Deleted || t.Archived || t.DateClosedAt != nil
}

func (t *Task) TotalEstimate() int64 {
	t.lazyLoadSubTasks()

	total := int64(0)
	if t.TimeEstimateMs != nil {
		total = *t.TimeEstimateMs
	}
	for idx := range t.SubTasks {
		if t.SubTasks[idx].TimeEstimateMs != nil {
			total += *t.SubTasks[idx].TimeEstimateMs
		}
	}
	return total
}
func (t *Task) GetList(ctx context.Context) *List {
	if t.List != nil {
		return t.List
	}
	t.List = t.List.NewModel().(*List)
	getRefDoc(ctx, t.ListRef, t.List)
	return t.List
}

func (t *Task) GetFolder(ctx context.Context) *Folder {
	if t.Folder != nil {
		return t.Folder
	}
	t.Folder = t.Folder.NewModel().(*Folder)
	getRefDoc(ctx, t.FolderRef, t.Folder)
	return t.Folder
}

func (t *Task) GetTeam(ctx context.Context) *Team {
	if t.Team != nil {
		return t.Team
	}
	t.Team = t.Team.NewModel().(*Team)
	getRefDoc(ctx, t.TeamRef, t.Team)
	return t.Team
}

func (t *Task) GetParentTask(ctx context.Context) *Task {
	if t.ParentTask != nil {
		return t.ParentTask
	}
	t.ParentTask = t.ParentTask.NewModel().(*Task)
	getRefDoc(ctx, t.ParentTaskRef, t.ParentTask)
	return t.ParentTask
}

func (t *Task) GetMember(ctx context.Context) *Member {
	if t.CreatorMember != nil {
		return t.CreatorMember
	}
	t.CreatorMember = t.CreatorMember.NewModel().(*Member)
	getRefDoc(ctx, t.CreatorMemberRef, t.CreatorMember)
	return t.CreatorMember
}

func (t *Task) GetAssignees() []*Member {
	t.lazyLoadAssignees()
	return t.Assignees
}

func (t *Task) GetSubTasks() []*Task {
	t.lazyLoadSubTasks()
	return t.SubTasks
}

func (t *Task) GetLinkedTasks() []*Task {
	t.lazyLoadLinkedTasks()
	return t.LinkedTasks
}

func (t *Task) AssignedByEmail(in string) bool {
	members := t.GetAssignees()
	for idx := range members {
		if strings.EqualFold(members[idx].Email, in) {
			return true
		}
	}
	return false
}

type List struct {
	StdStoreModel
	Name      string
	Archived  bool
	FolderRef *DocRef
	Folder    *Folder `firestore:"-"`
}

func (t *List) GetFolder(ctx context.Context) *Folder {
	if t.Folder != nil {
		return t.Folder
	}
	t.Folder = t.Folder.NewModel().(*Folder)
	getRefDoc(ctx, t.FolderRef, t.Folder)
	return t.Folder
}

type Folder struct {
	StdStoreModel
	Name     string
	Archived bool
}

type Team struct {
	StdStoreModel
	Name string
}

type Member struct {
	StdStoreModel
	Username string
	Email    string
	Initials string
}

func (*Task) NewModel() StoreModel {
	return &Task{}
}
func (t *Task) CollectionName() string {
	return "clickup_tasks"
}

func (t *List) CollectionName() string {
	return "clickup_lists"
}
func (*List) NewModel() StoreModel {
	return &List{}
}

func (t *Folder) CollectionName() string {
	return "clickup_folders"
}
func (*Folder) NewModel() StoreModel {
	return &Folder{}
}

func (t *Team) CollectionName() string {
	return "clickup_teams"
}
func (*Team) NewModel() StoreModel {
	return &Team{}
}

func (*Member) NewModel() StoreModel {
	return &Member{}
}
func (t *Member) CollectionName() string {
	return "clickup_members"
}
