package sync

import (
	"github.com/continuouspipe/remote-environment-client/sync/monitor"
	"github.com/continuouspipe/remote-environment-client/sync/rsync"
)

//syncs the files specified in filePaths. When filePaths is an empty slice, it syncs all project files
type Syncer interface {
	Sync(filePaths []string) error
	SetKubeConfigKey(string)
	SetEnvironment(string)
	SetPod(string)
	SetIndividualFileSyncThreshold(int)
	SetRemoteProjectPath(string)
	SetVerbose(bool)
}

func GetSyncer() Syncer {
	return rsync.GetRsync()
}

//this wraps a Syncer struct in order to implement the EventsObserver
//when a file changes OnLastChange() is called which will then trigger
//the sync via syncer.Sync
type SyncOnEvent struct {
	syncer Syncer
}

func NewSyncOnEvent() *SyncOnEvent {
	return &SyncOnEvent{}
}

func (observer SyncOnEvent) OnLastChange(filePaths []string) error {
	return observer.syncer.Sync(filePaths)
}

func GetSyncOnEventObserver(syncer Syncer) monitor.EventsObserver {
	syncOnEvent := NewSyncOnEvent()
	syncOnEvent.syncer = syncer
	return syncOnEvent
}
