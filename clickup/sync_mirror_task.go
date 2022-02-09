package clickup

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gebv/asap-tools/clickup/api"
	"go.uber.org/zap"
)

func MirrorTaskSyncer(api *api.API, store *Storage) *mirrorTaskSyncer {
	return &mirrorTaskSyncer{
		api:   api,
		store: store,
		log:   zap.L().Named("sync_mirror_task"),
	}
}

type mirrorTaskSyncer struct {
	api   *api.API
	store *Storage
	log   *zap.Logger
}

func (s *mirrorTaskSyncer) Sync(ctx context.Context, opts *SyncPreferences, oldTask, task *Task, changed bool) {
	if !changed {
		s.log.Debug("skip the not changed task", zap.String("task_id", task.ID))
		return
	}

	mirrorList, crossed := s.store.AllMatchesForMirrorTasks(ctx, task.ID)

	if !oldTask.Exists() && len(mirrorList) == 0 && task.IsDeletedOrHidden() {
		s.log.Debug("received the archived or closed task - not doing anything", zap.String("task_id", task.ID))
		return
	}

	// skip multi sync mode
	if crossed {
		s.log.Warn("multi-sync (when the task is both a mirror and a source) is not supported", zap.String("task_id", task.ID))
		return
	}

	rules := s.matchedRules(opts.MirrorTaskRules, task)

	// каждую MirrorTask обрабтать
	// - если это orig task то применить правило для orig task из mirror task
	// - если это mirror task %%%
	//
	// если среди всех зеркальныйх заданий текущая задача является исходной то
	// сохраняем listID в котром находится зеркальная задача (1)
	//
	// если входящая задача имеет listID отличный от (1) то обрабатываем как новую

	listOfMirrorTaskLists := map[string]bool{}
	for idx := range mirrorList {
		mirror := mirrorList[idx]
		if mirror.Destroyed {
			s.log.Warn("skipped the destroyed mirror-task", zap.String("mirror_task_id", mirror.ModelID()))
			continue
		}

		if mirror.TaskRef.ID == task.ID {
			for idx := range rules.changedRules {
				rule := rules.changedRules[idx]
				s.applyChangesToOriginalTask(ctx, mirror, rule, oldTask, task)
			}
		}

		if mirror.MirrorTaskRef.ID == task.ID {
			for idx := range rules.syncedRules {
				rule := rules.syncedRules[idx]
				s.applyChangesToMirrorTask(ctx, mirror, rule, oldTask, task)
			}
		}

		// если среди всех зеркальныйх заданий текущая задача является исходной то
		// сохраняем listID в котром находится зеркальная задача
		if task.ID == mirror.GetTask(ctx).ID {
			listOfMirrorTaskLists[mirror.GetMirrorTask(ctx).ListRef.ID] = true
		}
	}

	if !task.IsDeletedOrHidden() {
		for idx := range rules.addRules {
			rule := rules.addRules[idx]

			// если среди всех листов не встречается лист для правила добавления
			// тогда текущая задача кандидант на добавление в зеркало
			if !listOfMirrorTaskLists[rule.SpecAdd.GetAddToListID()] {
				s.addMirrorTask(ctx, rule.SpecAdd, task)
			}
		}
	}

}

func (s *mirrorTaskSyncer) destroyMirrorTask(ctx context.Context, mirror *MirrorTask, reason string) {
	mirror.Destroyed = true
	mirror.DestroyedAt = TimestampNow()
	mirror.DestroyedReason = reason
	err := s.store.UpsertMirrorTask(ctx, mirror)
	warnErrorIf(s.log, err, "failed destroy mirror task", "model_id", mirror.ModelID())
}

func (s *mirrorTaskSyncer) applyChangesToOriginalTask(ctx context.Context, mirror *MirrorTask,
	spec MirrorTaskSpecification, oldTask, task *Task) {

	if oldTask == nil {
		s.log.Error("handle task for orig task - old task was nil (not happen)", zap.String("task_id", task.ID))
		return
	}

	if mirror.GetMirrorTask(ctx).IsDeletedOrHidden() {
		s.sendComment(ctx, mirror.MirrorTaskRef.ID, "UNLINK MIRROR TASK: the mirror task has been DELETED or HIDDEN",
			spec.CondAdd.IfAssignedToMemberEmail)
		s.destroyMirrorTask(ctx, mirror, "mirror task has been DELETED or HIDDEN")
		return
	}

	commentText := &bytes.Buffer{}
	fmt.Fprintln(commentText, "The original task has chaned or differences with the mirror task:")
	needToUpdateTask := false
	needToSendComment := false
	updTask := &api.UpdateTaskRequest{
		TaskID: mirror.MirrorTaskRef.ID,
	}

	// track task name changes
	if oldTask.Name != task.Name {
		// fmt.Fprintf(commentText, "- changed name from %q to %q", oldTask.Name, task.Name)
		needToUpdateTask = true
		updTask.Name = task.MirrorTaskName(ctx)
	}
	// track task description changes
	if oldTask.Description != task.Description {
		needToUpdateTask = true
		updTask.Description = task.MirrorTaskDescription()
	}
	// track priority changes
	origPriorityID := task.PriorityID
	mirrorPriorityID := mirror.GetMirrorTask(ctx).PriorityID
	if origPriorityID != nil && mirrorPriorityID == nil {
		needToUpdateTask = true
		updTask.Priority = origPriorityID
	}
	if origPriorityID != nil && mirrorPriorityID != nil &&
		origPriorityID != mirrorPriorityID {
		needToUpdateTask = true
		updTask.Priority = origPriorityID
	}
	if origPriorityID == nil && mirrorPriorityID != nil {
		needToUpdateTask = true
		zero := 0
		updTask.Priority = &zero
	}

	// track task status changes
	if oldTask.StatusName != task.StatusName {
		fmt.Fprintf(commentText, "- changed task status name from %q to %q\n", oldTask.StatusName, task.StatusName)
		needToSendComment = true
	}
	// track task clsed at changes
	if oldTask.DateClosedAt == nil && task.DateClosedAt != nil {
		fmt.Fprintf(commentText, "- closed\n")
		needToSendComment = true
	}
	if oldTask.Archived == false && oldTask.Archived == true {
		fmt.Fprintf(commentText, "- archived\n")
		needToSendComment = true
	}
	if oldTask.Deleted == false && oldTask.Deleted == true {
		fmt.Fprintf(commentText, "- deleted\n")
		needToSendComment = true
	}

	// track location of the task by folder
	if oldTask.FolderRef.ID != task.FolderRef.ID {
		fmt.Fprintf(commentText, "- moved to the folder %q\n", task.GetFolder(ctx).Name)
		needToSendComment = true
	}
	// track location of the task by list
	if oldTask.ListRef.ID != task.ListRef.ID {
		fmt.Fprintf(commentText, "- moved to the list %q\n", task.GetList(ctx).Name)
		needToSendComment = true
	}

	// track task estimate changes
	totalEstimate := task.TotalEstimate()
	miirorTotalEsimate := mirror.GetMirrorTask(ctx).TotalEstimate()
	if miirorTotalEsimate > 0 && totalEstimate == 0 {
		// removed time estimate
		fmt.Fprintf(commentText, "- different time estimate - should be equal to %q but nil\n", msHuman(miirorTotalEsimate))
		needToSendComment = true
	}
	if totalEstimate > 0 && miirorTotalEsimate > 0 && miirorTotalEsimate != totalEstimate {
		// changed estimate from to
		fmt.Fprintf(commentText, "- different time estimate  - equals %s but should be equal to %s\n", msHuman(totalEstimate), msHuman(miirorTotalEsimate))
		needToSendComment = true
	}
	if miirorTotalEsimate == 0 && totalEstimate > 0 {
		// added estimate
		fmt.Fprintf(commentText, "- different time estimate - should be nil but equals to %q\n", msHuman(totalEstimate))
		needToSendComment = true
	}

	// track task due date changes
	origDueDate := task.DueDateAt
	mirrorDueDate := mirror.GetMirrorTask(ctx).DueDateAt
	if origDueDate != nil && mirrorDueDate == nil {
		// removed duedate
		fmt.Fprintln(commentText, "- different due date - should be nil")
		needToSendComment = true
	}
	if origDueDate != nil && mirrorDueDate != nil &&
		(*origDueDate).AsTime().Unix() != (*mirrorDueDate).AsTime().Unix() {
		// changed duedate from to
		fmt.Fprintf(commentText, "- different due date  - equals %s but should be equal to %s\n",
			origDueDate.AsTime().Format(time.RFC3339),
			mirrorDueDate.AsTime().Format(time.RFC3339))
		needToSendComment = true
	}
	if origDueDate == nil && mirrorDueDate != nil {
		// added duedate
		fmt.Fprintln(commentText, "- different due date - should be equal to", mirrorDueDate.AsTime().Format(time.RFC3339))
		needToSendComment = true
	}

	origStartDate := task.StartDateAt
	mirrorStartDate := mirror.GetMirrorTask(ctx).StartDateAt
	if origStartDate != nil && mirrorStartDate == nil {
		// removed duedate
		fmt.Fprintln(commentText, "- different due date - should be nil")
		needToSendComment = true
	}
	if origStartDate != nil && mirrorStartDate != nil &&
		(*origStartDate).AsTime().Unix() != (*mirrorStartDate).AsTime().Unix() {
		// changed duedate from to
		fmt.Fprintf(commentText, "- different due date  - equals %s but should be equal to %s\n",
			origStartDate.AsTime().Format(time.RFC3339),
			mirrorStartDate.AsTime().Format(time.RFC3339))
		needToSendComment = true
	}
	if origStartDate == nil && mirrorStartDate != nil {
		// added duedate
		fmt.Fprintln(commentText, "- different due date - should be equal to", mirrorStartDate.AsTime().Format(time.RFC3339))
		needToSendComment = true
	}

	if mirror.GetMirrorTask(ctx).IsDeletedOrHidden() {
		fmt.Fprintln(commentText, "- mirror task is archived or closed but something has changed in original task", task.URL)
		needToSendComment = true
	}

	if needToUpdateTask {
		updatedTaskAPI := s.api.UpdateTask(ctx, updTask)
		warnIfFailedRequest(s.log, updatedTaskAPI)
		if updatedTaskAPI.StatusOK() {
			updatedTask := ModelTaskFromAPI(ctx, s.store, &updatedTaskAPI.Task)
			err := s.store.UpsertTask(ctx, updatedTask)
			warnErrorIf(s.log, err, "failed to update a mirror task after processing changes and apply changes", "task_id", updatedTask.ID)
		}
	}

	if needToSendComment {
		s.sendComment(ctx, mirror.MirrorTaskRef.ID, commentText.String(), spec.SpecAdd.AssignToMemberEmail)
	}
}

func (s *mirrorTaskSyncer) applyChangesToMirrorTask(ctx context.Context, mirror *MirrorTask,
	spec MirrorTaskSpecification, oldTask, task *Task) {

	if task.IsDeletedOrHidden() {
		s.sendComment(ctx, task.ID, "FYI changes have been made to a mirror task that is DELETED or HIDDEN - nothing will be updated in original tasks and UNLINK MIRROR TASK",
			spec.CondAdd.IfAssignedToMemberEmail)
		s.destroyMirrorTask(ctx, mirror, "mirror task has been DELETED or HIDDEN")
		return
	}

	if mirror.GetTask(ctx).IsDeletedOrHidden() {
		s.sendComment(ctx, task.ID, "UNLINK MIRROR TASK: the original task has been DELETED or HIDDEN",
			spec.CondAdd.IfAssignedToMemberEmail)
		s.destroyMirrorTask(ctx, mirror, "original task has been DELETED or HIDDEN")
		return
	}

	origTask := mirror.GetTask(ctx)
	if !origTask.Exists() {
		s.log.Warn("handle task for mirror task - orig task was nil (why?)", zap.String("orig_task_id", mirror.TaskRef.ID),
			zap.String("mirror_task_id", mirror.TaskRef.ID),
			zap.String("task_id", task.ID),
		)
		return
	}

	commentText := &bytes.Buffer{}
	fmt.Fprintln(commentText, "The mirror task has chaned or differences with the original task (by main fields - estimate, due date, start date):")
	needToUpdateTask := false
	needToSendComment := false
	updTask := &api.UpdateTaskRequest{
		TaskID: mirror.TaskRef.ID,
	}
	updMirrorTask := &api.UpdateTaskRequest{
		TaskID: mirror.MirrorTaskRef.ID,
	}
	needToUpdateMirrorTask := false

	// task name
	mirrorTaskName := origTask.MirrorTaskName(ctx)
	if mirrorTaskName != task.Name {
		needToUpdateMirrorTask = true
		updMirrorTask.Name = mirrorTaskName
	}

	// visibility status
	if mirror.GetMirrorTask(ctx).IsDeletedOrHidden() && !mirror.GetTask(ctx).IsDeletedOrHidden() {
		fmt.Fprintf(commentText, "- mirror task has removed. TODO: remove from the mirror?\n")
		needToSendComment = true

	}

	// TODO: description change

	// estimate
	totalEstimate := task.TotalEstimate()
	if origTask.TimeEstimateMs != nil && totalEstimate == 0 {
		// removed time estimate
		needToUpdateTask = true
		updTask.TimeEstimateMs = -1
		fmt.Fprintf(commentText, "- orig task will be updated - time estimate will be removed\n")
		needToSendComment = true
	} else if origTask.TimeEstimateMs != nil && totalEstimate != 0 && totalEstimate != *origTask.TimeEstimateMs {
		// changed estimate from to
		needToUpdateTask = true
		updTask.TimeEstimateMs = totalEstimate
		fmt.Fprintf(commentText, "- orig task will be updated - time estimate will be sets to %s\n", msHuman(totalEstimate))
		needToSendComment = true
	}
	if origTask.TimeEstimateMs == nil && totalEstimate > 0 {
		// added estimate
		needToUpdateTask = true
		updTask.TimeEstimateMs = totalEstimate
		fmt.Fprintf(commentText, "- orig task will be updated - time estimate will be sets to %s\n", msHuman(totalEstimate))
		needToSendComment = true
	}

	// due date
	if origTask.DueDateAt != nil && task.DueDateAt == nil {
		// removed duedate
		needToUpdateTask = true
		updTask.DueDate = -1
		fmt.Fprintf(commentText, "- orig task will be updated - time due date will be removed\n")
		needToSendComment = true
	}
	if origTask.DueDateAt != nil && task.DueDateAt != nil &&
		(*origTask.DueDateAt).AsTime().Unix() != (*task.DueDateAt).AsTime().Unix() {
		// changed duedate from to
		needToUpdateTask = true
		updTask.DueDate = (*task.DueDateAt).AsTime().Unix() * 1000
		fmt.Fprintf(commentText, "- orig task will be updated - time due date will be sets to %s\n", (*task.DueDateAt).AsTime().Format(time.RFC3339))
		needToSendComment = true
	}
	if origTask.DueDateAt == nil && task.DueDateAt != nil {
		// added duedate
		needToUpdateTask = true
		updTask.DueDate = (*task.DueDateAt).AsTime().Unix() * 1000
		fmt.Fprintf(commentText, "- orig task will be updated - time due date will be sets to %s\n", (*task.DueDateAt).AsTime().Format(time.RFC3339))
		needToSendComment = true
	}

	// start date
	if origTask.StartDateAt != nil && task.StartDateAt == nil {
		// removed startdate
		needToUpdateTask = true
		updTask.StartDate = -1
		fmt.Fprintf(commentText, "- orig task will be updated - time start date will be removed\n")
		needToSendComment = true
	}
	if origTask.StartDateAt != nil && task.StartDateAt != nil &&
		(*origTask.StartDateAt).AsTime().Unix() != (*task.StartDateAt).AsTime().Unix() {
		// changed startdate from to
		needToUpdateTask = true
		updTask.StartDate = (*task.StartDateAt).AsTime().Unix() * 1000
		fmt.Fprintf(commentText, "- orig task will be updated - time start date will be sets to %s\n", (*task.StartDateAt).AsTime().Format(time.RFC3339))
		needToSendComment = true
	}
	if origTask.StartDateAt == nil && task.StartDateAt != nil {
		// added startdate
		needToUpdateTask = true
		updTask.StartDate = (*task.StartDateAt).AsTime().Unix() * 1000
		fmt.Fprintf(commentText, "- orig task will be updated - time start date will be sets to %s\n", (*task.StartDateAt).AsTime().Format(time.RFC3339))
		needToSendComment = true
	}

	if needToUpdateMirrorTask {
		updatedTaskAPI := s.api.UpdateTask(ctx, updMirrorTask)
		warnIfFailedRequest(s.log, updatedTaskAPI)

		if updatedTaskAPI.StatusOK() {
			updatedTask := ModelTaskFromAPI(ctx, s.store, &updatedTaskAPI.Task)
			err := s.store.UpsertTask(ctx, updatedTask)
			warnErrorIf(s.log, err, "failed to update a mirror task after processing changes and apply changes", "task_id", updatedTask.ID)
		}
	}

	if needToUpdateTask {
		updatedTaskAPI := s.api.UpdateTask(ctx, updTask)
		warnIfFailedRequest(s.log, updatedTaskAPI)

		if updatedTaskAPI.StatusOK() {
			updatedTask := ModelTaskFromAPI(ctx, s.store, &updatedTaskAPI.Task)
			err := s.store.UpsertTask(ctx, updatedTask)
			warnErrorIf(s.log, err, "failed to update a original task after processing changes and apply changes", "task_id", updatedTask.ID)
		}
	}

	if needToSendComment {
		s.sendComment(ctx, mirror.MirrorTaskRef.ID, commentText.String(), spec.SpecAdd.AssignToMemberEmail)
	}
}

func (s *mirrorTaskSyncer) addMirrorTask(ctx context.Context, spec *SyncRule_SpecOfAdd, task *Task) {
	taskID := task.ID
	l := s.log.Named("add_mirror_task").With(zap.String("task_id", taskID))

	if task.IsDeletedOrHidden() {
		l.Debug("aborted - task was deleted or hidden")
		return
	}

	mirrorTask := &api.CreateTaskRequest{
		ListID:              spec.GetAddToListID(),
		Name:                task.MirrorTaskName(ctx),
		RefTaskID:           task.ID,
		Tags:                []string{"mirror"},
		DescriptionMarkdown: task.MirrorTaskDescription(),
	}
	if spec.SetStatusName != "" {
		mirrorTask.StatusName = spec.SetStatusName
	}

	commentText := &bytes.Buffer{}
	fmt.Fprintln(commentText, "The mirror task from "+task.URL)

	if spec.AssignToMemberEmail != "" {
		member := s.store.MemberByEmail(ctx, spec.AssignToMemberEmail)
		if member.Exists() {
			mirrorTask.AssignIDs = []string{member.ID}
		} else {
			l.Warn("failed find member by email", zap.String("email", spec.AssignToMemberEmail))
			fmt.Fprintln(commentText, "Must be assigned to "+spec.AssignToMemberEmail)
		}
	}

	res := s.api.CreateTask(ctx, mirrorTask)
	warnIfFailedRequest(s.log, res)
	if res.TaskID == "" {
		l.Error("aborted creation of a mirror task - no ID from a new mirror task (from API)")
		return
	}
	err := s.store.UpsertMirrorTask(ctx, s.store.ModelMirrorTaskFor(taskID, res.TaskID))
	if err != nil {
		l.Error("failed add mirror task to database", zap.Error(err), zap.String("mirror_task_id", res.TaskID))
		return
	}

	fmt.Fprintln(commentText, "A ready go.")
	fmt.Fprintln(commentText)
	fmt.Fprintln(commentText, "NOTES: ")
	fmt.Fprintln(commentText, "- description without markdown formatting")
	fmt.Fprintln(commentText, "- sets estimate in subtasks do not initiate a push into the original task (2022/01/18)")
	fmt.Fprintln(commentText, "- not always the due date is pushed into the orig task")
	s.sendComment(ctx, res.TaskID, commentText.String(), "")
}

func (s *mirrorTaskSyncer) sendComment(ctx context.Context, taskID string, commentText string, assignToEmail string) {
	comment := &api.AddCommentToTaskRequest{
		TaskID:      taskID,
		CommentText: commentText,
	}
	if assignToEmail != "" {
		member := s.store.MemberByEmail(ctx, assignToEmail)
		if member.Exists() {
			comment.AssignToMemberID = fmt.Sprint(member.ID)
		} else {
			s.log.Warn("comment sending - not found member by email", zap.String("email", assignToEmail),
				zap.String("task_id", taskID))
		}
	}
	res := s.api.AddCommentToTask(ctx, comment)
	warnIfFailedRequest(s.log, res)
}

type syncMirrorTasksMatchedRules struct {
	task                                *Task
	addRules, changedRules, syncedRules []MirrorTaskSpecification
}

// returns the rules that match the task
func (r *mirrorTaskSyncer) matchedRules(rules []MirrorTaskSpecification, task *Task) *syncMirrorTasksMatchedRules {
	res := &syncMirrorTasksMatchedRules{task: task}

	if len(rules) == 0 {
		return res
	}

	// teamID := task.TeamRef.ID
	folderID := task.FolderRef.ID
	listID := task.ListRef.ID
	status := task.StatusName

	// if incoming task the candidate for the sync
	for idx := range rules {
		rule := rules[idx]
		if !rule.existsRultesForTeamID(task.TeamID()) {
			// skiped, because no rules for task
			continue
		}

		// checking the cond for sync of the task
		if cond := rule.CondAdd; cond != nil {
			if cond.PassedCheckByFolderID(folderID) &&
				cond.PassedCheckByListID(listID) &&
				cond.PassedCheckByStatus(status) {

				if cond.IfAssignedToMemberEmail == "" ||
					(cond.IfAssignedToMemberEmail != "" && task.AssignedByEmail(cond.IfAssignedToMemberEmail)) {
					res.addRules = append(res.addRules, rule)
				}
			}
		}

		// checking the cond for track changes of the tasks
		if cond := rule.CondTrackChanges; cond != nil {
			if cond.PassedCheckByFolderID(folderID) &&
				cond.PassedCheckByListID(listID) &&
				cond.PassedCheckByStatus(status) {

				if cond.IfAssignedToMemberEmail == "" ||
					(cond.IfAssignedToMemberEmail != "" && task.AssignedByEmail(cond.IfAssignedToMemberEmail)) {
					res.changedRules = append(res.changedRules, rule)
				}
			}
		}
	}

	for idx := range rules {
		rule := rules[idx]

		if spec := rule.SpecAdd; spec != nil {
			if spec.GetAddToListID() == listID {
				res.syncedRules = append(res.syncedRules, rule)
			}
		}
	}

	zap.L().Debug("result of the parsing of the rules on the task", zap.String("task_id", task.ID),
		zap.Int("num_changed_rules", len(res.changedRules)),
		zap.Int("num_synced_rules", len(res.syncedRules)),
		zap.Int("num_added_rules", len(res.addRules)))

	return res
}

type MirrorTaskSpecification struct {
	Name string `yaml:"name"`

	CondAdd          *SyncRule_CondOfAdd        `yaml:"cond_add"`
	CondTrackChanges *SyncRule_CondTrackChanges `yaml:"cond_track_changes"`
	SpecAdd          *SyncRule_SpecOfAdd        `yaml:"spec_add"`
}

func (r *MirrorTaskSpecification) existsRultesForTeamID(teamID string) bool {
	for _, v := range r.UsedTeamIDs() {
		if v == teamID {
			return true
		}
	}
	return false
}

func (r *MirrorTaskSpecification) UsedTeamIDs() []string {
	uniq := map[string]bool{}
	res := []string{}
	for _, folderURL := range r.CondAdd.IfInFolders {
		id := teamIDFromURL(folderURL)
		if !uniq[id] {
			uniq[id] = true
			res = append(res, id)
		}
	}
	for _, listURL := range r.CondAdd.IfInLists {
		id := teamIDFromURL(listURL)
		if !uniq[id] {
			uniq[id] = true
			res = append(res, id)
		}
	}
	for _, folderURL := range r.CondTrackChanges.IfInFolders {
		id := teamIDFromURL(folderURL)
		if !uniq[id] {
			uniq[id] = true
			res = append(res, id)
		}
	}
	for _, folderURL := range r.CondTrackChanges.IfInLists {
		id := teamIDFromURL(folderURL)
		if !uniq[id] {
			uniq[id] = true
			res = append(res, id)
		}
	}

	id := teamIDFromURL(r.SpecAdd.AddToList)
	if !uniq[id] {
		uniq[id] = true
		res = append(res, id)
	}

	return res
}

type SyncRule_CondOfAdd struct {
	// Example: https://app.clickup.com/2431928/v/f/96471870/42552884
	IfInFolders []string `yaml:"if_in_folders"`
	// Example: https://app.clickup.com/2431928/v/li/174318787
	IfInLists []string `yaml:"if_in_lists"`

	EqAnyTaskStatusNames    []string `yaml:"eq_any_task_status_names"`
	IfAssignedToMemberEmail string   `yaml:"if_assigned_to_member_email"`
	// TODO: add more flexible conds
}

type SyncRule_CondTrackChanges struct {
	// Example: https://app.clickup.com/2431928/v/f/96471870/42552884
	IfInFolders []string `yaml:"if_in_folders"`
	// Example: https://app.clickup.com/2431928/v/li/174318787
	IfInLists []string `yaml:"if_in_lists"`

	EqAnyTaskStatusNames    []string `yaml:"eq_any_task_status_names"`
	IfAssignedToMemberEmail string   `yaml:"if_assigned_to_member_email"`
	// TODO: add more flexible conds
}

type SyncRule_SpecOfAdd struct {
	// synced tasks to list
	AddToList           string `yaml:"add_to_list"`
	SetStatusName       string `yaml:"set_status_name"`
	AssignToMemberEmail string `yaml:"assign_to_member_email"`
	// TODO: add more flexible rules
	// For eg.
	// - send comment?
	// - add tag?
}

func (s *SyncRule_SpecOfAdd) GetAddToListID() string {
	return listIDFromURL(s.AddToList)
}

func (s *SyncRule_CondOfAdd) PassedCheckByStatus(in string) bool {
	if len(s.EqAnyTaskStatusNames) == 0 {
		return true
	}
	for _, status := range s.EqAnyTaskStatusNames {
		if strings.EqualFold(status, in) {
			return true
		}
	}
	return false
}

func (s *SyncRule_CondTrackChanges) PassedCheckByStatus(in string) bool {
	if len(s.EqAnyTaskStatusNames) == 0 {
		return true
	}
	for _, status := range s.EqAnyTaskStatusNames {
		if strings.EqualFold(status, in) {
			return true
		}
	}
	return false
}

func (s *SyncRule_CondOfAdd) PassedCheckByFolderID(in string) bool {
	if len(s.GetIfInFolderIDs()) == 0 {
		return true
	}

	for _, folderID := range s.GetIfInFolderIDs() {
		if folderID == in {
			return true
		}
	}
	return false
}

func (s *SyncRule_CondOfAdd) PassedCheckByListID(in string) bool {
	if len(s.GetIfInListsIDs()) == 0 {
		return true
	}

	for _, folderID := range s.GetIfInListsIDs() {
		if folderID == in {
			return true
		}
	}
	return false
}

func (s *SyncRule_CondOfAdd) GetIfInFolderIDs() []string {
	res := []string{}
	for _, folderURL := range s.IfInFolders {
		folderID := folderIDFromURL(folderURL)
		if folderID == "" {
			continue
		}
		res = append(res, folderID)
	}
	return res
}

func (s *SyncRule_CondOfAdd) GetIfInListsIDs() []string {
	res := []string{}
	for _, listID := range s.IfInLists {
		listID := listIDFromURL(listID)
		if listID == "" {
			continue
		}
		res = append(res, listID)
	}
	return res
}

func (s *SyncRule_CondTrackChanges) GetIfInFolderIDs() []string {
	res := []string{}
	for _, folderURL := range s.IfInFolders {
		folderID := folderIDFromURL(folderURL)
		if folderID == "" {
			continue
		}
		res = append(res, folderID)
	}
	return res
}

func (s *SyncRule_CondTrackChanges) GetIfInListsIDs() []string {
	res := []string{}
	for _, listID := range s.IfInLists {
		listID := listIDFromURL(listID)
		if listID == "" {
			continue
		}
		res = append(res, listID)
	}
	return res
}

func (s *SyncRule_CondTrackChanges) PassedCheckByFolderID(in string) bool {
	if len(s.GetIfInFolderIDs()) == 0 {
		return true
	}
	for _, folderID := range s.GetIfInFolderIDs() {
		if folderID == in {
			return true
		}
	}
	return false
}

func (s *SyncRule_CondTrackChanges) PassedCheckByListID(in string) bool {
	if len(s.GetIfInListsIDs()) == 0 {
		return true
	}
	for _, folderID := range s.GetIfInListsIDs() {
		if folderID == in {
			return true
		}
	}
	return false
}
