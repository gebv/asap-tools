package clickup

import "context"

type taskSyncer interface {
	Sync(ctx context.Context, opts *SyncPreferences, oldTask, task *Task, changed bool)
}

var _ taskSyncer = (*ChangeManager)(nil)
var _ taskSyncer = (*mirrorTaskSyncer)(nil)
