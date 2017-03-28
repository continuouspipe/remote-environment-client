package rsync

import "runtime"

//rsync exclusion file used when fetching and syncing
const SyncFetchExcluded = ".cp-remote-ignore"
//rsync exclusion file used only when fetching
const FetchExcluded = ".cp-remote-ignore-fetch"


//use rsync to sync the files specified in filePaths. When filePaths is an empty slice, it syncs all project files
type RsyncSyncer interface {
	Sync(paths []string) error
	SetKubeConfigKey(string)
	SetEnvironment(string)
	SetPod(string)
	SetIndividualFileSyncThreshold(int)
	SetRemoteProjectPath(string)
	SetVerbose(bool)
}

var RsyncRsh RsyncSyncer
var RsyncDaemon RsyncSyncer

func GetRsync() RsyncSyncer {
	if runtime.GOOS == "windows" {
		return RsyncDaemon
	}
	return RsyncRsh
}
