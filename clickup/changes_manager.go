package clickup

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gebv/asap-tools/clickup/api"
	"go.uber.org/zap"
)

func NewChangeManager(api *api.API, s *Storage) *ChangeManager {
	return &ChangeManager{
		api:   api,
		store: s,
		log:   zap.L().Named("clickup_sync"),
	}
}

type ChangeManager struct {
	api           *api.API
	store         *Storage
	log           *zap.Logger
	webhookSecret string
}

func (s *ChangeManager) Sync(ctx context.Context, opts *SyncPreferences, oldTask, task *Task, changed bool) {
	list := []taskSyncer{
		MirrorTaskSyncer(s.api, s.store),
	}

	for _, syncer := range list {
		// TODO: models must not be modified
		syncer.Sync(ctx, opts, oldTask, task, changed)
	}
}

// ForceSyncForAllTasks force update each task from the database and apply processing to it.
// Not found tasks from ClickUp API to marked as deleted.
func (s *ChangeManager) ForceSyncForAllTasks(ctx context.Context, opts *SyncPreferences, teamID string) {
	list := s.store.AllTeamTasks(ctx, teamID)
	for idx := range list {
		oldTask := list[idx]
		if oldTask.Deleted {
			// TODO: add index into firestore and add deploy script
			continue
		}

		res := s.api.TaskByID(ctx, oldTask.ID)
		if res.StatusOK() {
			newTask := ModelTaskFromAPI(ctx, s.store, &res.Task)
			s.Sync(ctx, opts, oldTask, newTask, true)
		}

		// ClickUp API returns 404 if the task is not found
		if res.IsStatus(http.StatusNotFound) {
			// task was deleted
			oldTask.Deleted = true
			s.store.UpsertTask(ctx, oldTask)
			// TODO: add special logic for terminated tasks
		}

		if !res.StatusOK() && !res.IsStatus(http.StatusNotFound) {
			warnIfFailedRequest(s.log, res)
			s.log.Warn("failed to get task data from ClickUp API", zap.String("task_id", oldTask.ID))
		}
	}
}

// ApplyChangesInTeam fetchs the latest changed tasks and handle.
// Save the date of the last changed task.
//
// Processing details:
// - call to ClickUp API "give me changes tasks"
// - fetch and upsert related (list, folder, members) data if not exists
// - processing for each tasks
//   - lookup for rules to add mirror tasks and add if need
//   - lookup for rules to track changes for mirror tasks or for tasks that have a mirror task and process if need
func (s *ChangeManager) ApplyChangesInTeam(ctx context.Context, opts *SyncPreferences, teamID string) {
	l := s.log.Named("handle_latest_changes").With(zap.String("team_id", teamID))

	cursor := s.store.GetStateOfLoadChangesForTeamTasks(ctx, teamID)

	l.Debug("Loading from API latest changes tasks for team", zap.Any("latest_changes_in_team_cursor", cursor))

	page := 0
	lastTaskUpdatedAt := int64(0)
	loadedListDeps := map[string]bool{}

nextPage:
	req := &api.SearchTasksInTeamRequest{
		TeamID:          teamID,
		OrderBy:         "updated",
		IncludeClosed:   true,
		IncludeSubtasks: true,
		Page:            page,
	}
	if cursor.Exists() && cursor.LastTaskUpdatedAt > 0 {
		// NOTE: cursor (order_by=updated, date_updated_gt and fetch condition) should not change, only update the page number
		req.DateUpdatedGtTs = cursor.LastTaskUpdatedAt
	}

	// TODO: the archived and closed tasks is not fall in the /teams/<TeamID>/tasks selection
	res := s.api.SearchTasksInTeam(ctx, req)
	warnIfFailedRequest(s.log, res)
	if !res.StatusOK() {
		l.Warn("failed getting a task list from API", zap.Any("request_opts", req))
		return
	}

	l.Debug("[LIST_CHANGED_TASKS] got from API the team tasks", zap.Any("request_opts", req), zap.Int("num_tasks", len(res.Tasks)))

	for idx := range res.Tasks {
		taskAPI := res.Tasks[idx]

		// load members, folders from list if not previously loaded
		if !loadedListDeps[taskAPI.List.ID] {
			s.fetchAndUpdateListAndListRelatedData(ctx, taskAPI.List.ID)
			loadedListDeps[taskAPI.List.ID] = true
		}

		lastTaskUpdatedAt = maxInt64(taskAPI.DateUpdatedTs, lastTaskUpdatedAt)

		task := ModelTaskFromAPI(ctx, s.store, &taskAPI)

		oldTask, changed := s.AuthorizeTask(ctx, task)
		s.Sync(ctx, opts, oldTask, task, changed)
	}

	// saved the cursor if it has changed
	if cursor.Exists() && lastTaskUpdatedAt > cursor.LastTaskUpdatedAt {
		err := s.store.UpsertLoadStatusOfChangedTeamTasks(ctx, &LoadStatusOfChangedTeamTasks{
			TeamID:            teamID,
			TeamRef:           s.store.DocRef(NewWithID(TeamModel, teamID)),
			LastTaskUpdatedAt: lastTaskUpdatedAt,
		})
		s.warnErrorIf(err, "failed upsert status of loading tasks from the team", "team_id", teamID)
	}

	if len(res.Tasks) == 100 {
		page++
		goto nextPage
	}
}

func ModelTaskSetupLazyload(ctx context.Context, store *Storage, model *Task) {
	model.lazyLoadAssignees = func() {
		model.lazyLoadAssigneesOnce.Do(func() {
			model.Assignees = store.FetchListMembers(ctx, model.AssigneesRef)
		})
	}
	model.lazyLoadSubTasks = func() {
		model.lazyLoadSubTasksOnce.Do(func() {
			model.LinkedTasks = store.FetchListTasks(ctx, model.LinkedTasksRef)
		})
	}
	model.lazyLoadLinkedTasks = func() {
		model.lazyLoadLinkedTasksOnce.Do(func() {
			store.LoadSubTasks(ctx, model)
		})
	}
}

func ModelTaskFromAPI(ctx context.Context, store *Storage, taskAPI *api.Task) *Task {
	model := &Task{
		StdStoreModel:  NewWithID(TaskModel, taskAPI.ID).(*Task).StdStoreModel,
		Name:           taskAPI.Name,
		CustomID:       taskAPI.CustomID,
		Description:    taskAPI.Description,
		TextContent:    taskAPI.TextContent,
		StatusType:     taskAPI.Status.Type,
		StatusName:     taskAPI.Status.Status,
		DateCreatedAt:  TimestampFromTimestampWithMilliseconds(&taskAPI.DateCreatedTs),
		DateUpdatedAt:  TimestampFromTimestampWithMilliseconds(&taskAPI.DateUpdatedTs),
		DateClosedAt:   TimestampFromTimestampWithMilliseconds(taskAPI.DateClosedTs),
		DueDateAt:      TimestampFromTimestampWithMilliseconds(taskAPI.DueDate),
		StartDateAt:    TimestampFromTimestampWithMilliseconds(taskAPI.StartDate),
		TimeEstimateMs: taskAPI.TimeEstimateMs,
		Archived:       taskAPI.Archived,
		URL:            taskAPI.URL,
		Tags:           taskAPI.ListTags(),

		// TODO: Load from the API if there is not exists in the database?
		TeamRef: store.DocRef(NewWithID(TeamModel, taskAPI.TeamID)),
		ListRef: store.DocRef(NewWithID(ListModel, taskAPI.List.ID)),
		// SpaceRef:      store.DocRef(&Space{StorageModel: StorageModel{ID: taskAPI.Space.ID}}),
		FolderRef:        store.DocRef(NewWithID(FolderModel, taskAPI.Folder.ID)),
		CreatorMemberRef: store.DocRef(NewWithID(MemberModel, fmt.Sprint(taskAPI.Creator.ID))),
		AssigneesRef:     []*firestore.DocumentRef{},
		LinkedTasksRef:   []*firestore.DocumentRef{},
	}

	if taskAPI.Parent != nil {
		model.ParentTaskRef = store.DocRef(NewWithID(TaskModel, *taskAPI.Parent))
	}
	if taskAPI.Priority != nil {
		model.PriorityID = &taskAPI.Priority.ID
	}

	for _, taskID := range taskAPI.ListLinkedTaskIDs() {
		model.LinkedTasksRef = append(model.LinkedTasksRef,
			store.DocRef(NewWithID(TaskModel, taskID)),
		)
	}
	for idx := range taskAPI.Assignees {
		assign := taskAPI.Assignees[idx]

		model.AssigneesRef = append(model.AssigneesRef,
			store.DocRef(NewWithID(MemberModel, fmt.Sprint(assign.ID))),
		)
	}

	ModelTaskSetupLazyload(ctx, store, model)

	return model
}

func (s *ChangeManager) AuthorizeTask(ctx context.Context, task *Task) (_ *Task, changed bool) {
	oldTask := s.store.GetTask(ctx, task.ID)

	changed = false

	if !oldTask.Exists() {
		s.log.Debug("[AUTH_NOT_FOUND] task has not found (will be added in database)",
			zap.String("task_id", task.ID),
			zap.Time("task_updated_at", task.DateUpdatedAt.AsTime()))

		// Новая задача - сохраняем в БД
		err := s.store.UpsertTask(ctx, task)
		s.warnErrorIf(err, "failed to upsert not exists task", "task_id", task.ID)

		changed = true
	}

	if oldTask.Exists() && oldTask.EqualUpdatedAt(task.DateUpdatedAt) {
		s.log.Debug("[AUTH_NO_CHANGES] task has not changes (will not be updated in database)",
			zap.String("task_id", task.ID),
			zap.Time("task_updated_at", task.DateUpdatedAt.AsTime()))

		// changed = false
		return oldTask, false
	}

	if oldTask.Exists() && !oldTask.EqualUpdatedAt(task.DateUpdatedAt) {
		s.log.Debug("[AUTH_CHANGES] task has changes (will be updated in database)",
			zap.String("task_id", task.ID),
			zap.Time("old_task_updated_at", oldTask.DateUpdatedAt.AsTime()),
			zap.Time("task_updated_at", task.DateUpdatedAt.AsTime()),
		)

		err := s.store.UpsertTask(ctx, task)
		s.warnErrorIf(err, "failed to upsert the changed task",
			"task_id", task.ID,
			"old_task_updated_at", oldTask.DateUpdatedAt.AsTime(),
			"task_updated_at", task.DateUpdatedAt.AsTime(),
		)

		changed = true
	}

	return oldTask, changed
}

// updates the list, the members in the list, folder in which is located
func (s *ChangeManager) fetchAndUpdateListAndListRelatedData(ctx context.Context, listID string) *List {
	list := s.fetchAndUpsertListIfNotExists(ctx, listID)

	s.fetchAndUpsertListMembersIfNotExists(ctx, listID, false)

	if list.FolderRef != nil && list.FolderRef.ID != "" {
		s.fetchAndUpsertFolderIfNotExists(ctx, list.FolderRef.ID)
	}
	if list.FolderRef == nil || list.FolderRef.ID == "" {
		s.warnErrorIf(fmt.Errorf("got an empty folder for the list"), "list_id", listID)
	}

	return list
}

func (s *ChangeManager) fetchAndUpsertListIfNotExists(ctx context.Context, modelID string) *List {
	model := s.store.GetList(ctx, modelID)
	if model.Exists() {
		return model
	}

	modelAPI := s.api.ListByID(ctx, modelID)
	warnIfFailedRequest(s.log, modelAPI)

	if !modelAPI.StatusOK() {
		s.warnErrorIf(fmt.Errorf("no data from API"), "no list from API", "list_id", modelID)
		return model
	}

	model = NewWithID(ListModel, modelID).(*List)
	model.Name = modelAPI.Name
	model.Archived = modelAPI.Archived
	// NOTE: always exists folder (user or service)
	model.FolderRef = s.store.DocRef(NewWithID(FolderModel, modelAPI.Folder.ID))

	err := s.store.UpsertIfNotExists(ctx, model)
	s.warnErrorIf(err, "failed upsert list")

	return model
}

func (s *ChangeManager) fetchAndUpsertFolderIfNotExists(ctx context.Context, modelID string) *Folder {
	model := s.store.GetFolder(ctx, modelID)
	if model.Exists() {
		return model
	}

	modelAPI := s.api.ListByID(ctx, modelID)
	warnIfFailedRequest(s.log, modelAPI)
	if !modelAPI.StatusOK() {
		s.warnErrorIf(fmt.Errorf("no data from API"), "no folder from API", "list_id", modelID)
		return model
	}

	model = NewWithID(FolderModel, modelID).(*Folder)
	model.Name = modelAPI.Name
	model.Archived = modelAPI.Archived

	err := s.store.UpsertIfNotExists(ctx, model)
	s.warnErrorIf(err, "failed upsert folder")

	return model
}

func (s *ChangeManager) fetchAndUpsertListMembersIfNotExists(ctx context.Context, modelID string, updateIfChanges bool) []*Member {
	listAPI := s.api.ListMembersOfList(ctx, modelID)
	warnIfFailedRequest(s.log, listAPI)
	if !listAPI.StatusOK() {
		s.warnErrorIf(fmt.Errorf("no data from API"), "no members of list from API", "list_id", modelID)
		return nil
	}

	res := []*Member{}
	for idx := range listAPI.Members {
		modelAPI := listAPI.Members[idx]

		model := s.store.GetMember(ctx, modelAPI.IDString())
		found := model.Exists()

		if !found {
			model = NewWithID(MemberModel, modelAPI.IDString()).(*Member)
		}

		changes := model.Email != modelAPI.Email ||
			model.Username != modelAPI.Username ||
			model.Initials != modelAPI.Initials

		// update member if changes data or not found
		if (changes && updateIfChanges) || !found {

			model.Email = modelAPI.Email
			model.Username = modelAPI.Username
			model.Initials = modelAPI.Initials

			err := s.store.UpsertMember(ctx, model)
			s.warnErrorIf(err, "failed upsert member", "member_id", modelAPI.IDString())
		}

		res = append(res, model)
	}
	return res
}

func (s *ChangeManager) warnErrorIf(err error, msg string, pairs ...interface{}) {
	if err != nil {
		s.log.Sugar().Warnw(msg, append([]interface{}{"err", err}, pairs...)...)
	}
}
