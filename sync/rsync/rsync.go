package rsync

import "runtime"

//use rsync to sync the files specified in filePaths. When filePaths is an empty slice, it syncs all project files
type RsyncSyncer interface {
	Sync(filePaths []string) error
	SetKubeConfigKey(string)
	SetEnvironment(string)
	SetPod(string)
	SetIndividualFileSyncThreshold(int)
}

var RsyncRsh RsyncSyncer
var RsyncDaemon RsyncSyncer

func GetRsync() RsyncSyncer {
	if runtime.GOOS == "windows" {
		return RsyncDaemon
	}
	return RsyncRsh
}
